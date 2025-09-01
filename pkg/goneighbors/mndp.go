package goneighbors

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const (
	MNDPPort         = 5678
	MNDPMulticast    = "255.255.255.255"
	MaxPacketSize    = 1500
	IPv6LinkLocalAll = "ff02::1"
)

// MNDPDiscoverer handles MikroTik Neighbor Discovery Protocol
type MNDPDiscoverer struct {
	udp4Conn *net.UDPConn
	udp6Conn *net.UDPConn
	running  bool
	stopChan chan struct{}
}

// NewMNDPDiscoverer creates a new MNDP discoverer
func NewMNDPDiscoverer() (*MNDPDiscoverer, error) {
	// Create UDP4 listener
	udp4Addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", MNDPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP4 address: %v", err)
	}

	udp4Conn, err := net.ListenUDP("udp4", udp4Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP4: %v", err)
	}

	// Create UDP6 listener (optional, may fail on some systems)
	udp6Addr, err := net.ResolveUDPAddr("udp6", fmt.Sprintf("[::]:%d", MNDPPort))
	var udp6Conn *net.UDPConn
	if err == nil {
		udp6Conn, _ = net.ListenUDP("udp6", udp6Addr)
	}

	return &MNDPDiscoverer{
		udp4Conn: udp4Conn,
		udp6Conn: udp6Conn,
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins listening for MNDP packets
func (d *MNDPDiscoverer) Start(neighborChan chan<- Neighbor) error {
	d.running = true

	// Start UDP4 listener
	go d.listenUDP4(neighborChan)

	// Start UDP6 listener if available
	if d.udp6Conn != nil {
		go d.listenUDP6(neighborChan)
	}

	return nil
}

// Stop stops the MNDP discoverer
func (d *MNDPDiscoverer) Stop() {
	if !d.running {
		return // Already stopped
	}

	d.running = false

	// Safely close the channel only if it's not already closed
	select {
	case <-d.stopChan:
		// Channel is already closed
	default:
		close(d.stopChan)
	}

	if d.udp4Conn != nil {
		d.udp4Conn.Close()
	}
	if d.udp6Conn != nil {
		d.udp6Conn.Close()
	}
}

// listenUDP4 listens for MNDP packets on UDP4
func (d *MNDPDiscoverer) listenUDP4(neighborChan chan<- Neighbor) {
	buffer := make([]byte, MaxPacketSize)

	for d.running {
		select {
		case <-d.stopChan:
			return
		default:
			d.udp4Conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := d.udp4Conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if d.running {
					fmt.Printf("Error reading UDP4 packet: %v\n", err)
				}
				continue
			}

			neighbor := d.parseMNDPPacket(buffer[:n], addr)
			if neighbor != nil {
				neighborChan <- *neighbor
			}
		}
	}
}

// listenUDP6 listens for MNDP packets on UDP6
func (d *MNDPDiscoverer) listenUDP6(neighborChan chan<- Neighbor) {
	buffer := make([]byte, MaxPacketSize)

	for d.running {
		select {
		case <-d.stopChan:
			return
		default:
			d.udp6Conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := d.udp6Conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if d.running {
					fmt.Printf("Error reading UDP6 packet: %v\n", err)
				}
				continue
			}

			neighbor := d.parseMNDPPacket(buffer[:n], addr)
			if neighbor != nil {
				neighborChan <- *neighbor
			}
		}
	}
}

// parseMNDPPacket parses an MNDP packet and returns a Neighbor
func (d *MNDPDiscoverer) parseMNDPPacket(data []byte, sourceAddr *net.UDPAddr) *Neighbor {
	if len(data) < 4 {
		return nil
	}

	reader := bytes.NewReader(data)
	msg := d.readMNDPMessage(reader)
	if msg == nil {
		return nil
	}

	neighbor := &Neighbor{
		IPAddress:    sourceAddr.IP.String(),
		Protocol:     ProtocolMNDP,
		SSHPort:      22,
		HasSSH:       true,
		HasTelnet:    true,
		DiscoveredAt: time.Now(),
		Platform:     "MikroTik",
	}

	// Process TLV fields
	for tag, tlv := range msg.Fields {
		d.processTLV(tag, tlv.Value, neighbor)
	}

	return neighbor
}

// readMNDPMessage reads an MNDP message from a reader
func (d *MNDPDiscoverer) readMNDPMessage(r io.Reader) *MNDPPacket {
	var msg MNDPPacket
	msg.Fields = make(map[TLVTag]TLV)

	// Read TypeTag
	err := binary.Read(r, binary.BigEndian, &msg.TypeTag)
	if err != nil {
		return nil
	}

	// Read SeqNo
	err = binary.Read(r, binary.BigEndian, &msg.SeqNo)
	if err != nil {
		return nil
	}

	// Read TLVs
	for {
		tlv := d.readTLV(r)
		if tlv == nil {
			break
		}
		msg.Fields[tlv.Tag] = *tlv
	}

	return &msg
}

// readTLV reads a TLV from a reader
func (d *MNDPDiscoverer) readTLV(r io.Reader) *TLV {
	var tag uint16
	var length uint16

	err := binary.Read(r, binary.BigEndian, &tag)
	if err != nil {
		return nil
	}

	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil
	}

	value := make([]byte, length)
	_, err = io.ReadFull(r, value)
	if err != nil {
		return nil
	}

	return &TLV{
		Tag:    TLVTag(tag),
		Length: length,
		Value:  value,
	}
}

// processTLV processes a TLV and updates the neighbor information
func (d *MNDPDiscoverer) processTLV(tag TLVTag, value []byte, neighbor *Neighbor) {
	switch tag {
	case TagMACAddr:
		if len(value) >= 6 {
			neighbor.MACAddress = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				value[0], value[1], value[2], value[3], value[4], value[5])
		}
	case TagIdentity:
		neighbor.Identity = strings.TrimRight(string(value), "\x00")
		neighbor.SystemName = neighbor.Identity
	case TagVersion:
		neighbor.Version = strings.TrimRight(string(value), "\x00")
		neighbor.SystemDesc = neighbor.Version
	case TagPlatform:
		neighbor.Platform = strings.TrimRight(string(value), "\x00")
	case TagUptime:
		if len(value) >= 4 {
			uptime := binary.BigEndian.Uint32(value)
			neighbor.Uptime = time.Duration(uptime) * time.Second
		}
	case TagSoftwareID:
		neighbor.SoftwareID = strings.TrimRight(string(value), "\x00")
	case TagBoard:
		neighbor.Board = strings.TrimRight(string(value), "\x00")
	case TagUnpack:
		neighbor.UnpackedData = strings.TrimRight(string(value), "\x00")
	case TagIPv6Addr:
		if len(value) >= 16 {
			neighbor.IPv6Address = net.IP(value[:16]).String()
		}
	case TagInterfaceName:
		neighbor.InterfaceName = strings.TrimRight(string(value), "\x00")
	case TagIPv4Addr:
		if len(value) >= 4 {
			ip := net.IPv4(value[0], value[1], value[2], value[3])
			// Use IPv4 address from TLV if available, otherwise use source address
			if !ip.IsUnspecified() {
				neighbor.IPAddress = ip.String()
			}
		}
	}
}

// SendMNDPDiscovery sends an MNDP discovery packet
func (d *MNDPDiscoverer) SendMNDPDiscovery() error {
	// Create MNDP discovery packet (TypeTag=0, SeqNo=0)
	packet := []byte{0x00, 0x00, 0x00, 0x00}

	// Send to IPv4 broadcast
	err := d.sendUDP4Broadcast(packet)
	if err != nil {
		fmt.Printf("Failed to send UDP4 broadcast: %v\n", err)
	}

	// Send to IPv6 multicast if available
	if d.udp6Conn != nil {
		err = d.sendUDP6Multicast(packet)
		if err != nil {
			fmt.Printf("Failed to send UDP6 multicast: %v\n", err)
		}
	}

	return nil
}

// sendUDP4Broadcast sends packet to all network broadcast addresses
func (d *MNDPDiscoverer) sendUDP4Broadcast(packet []byte) error {
	networks, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}

	for _, network := range networks {
		_, inet, err := net.ParseCIDR(network.String())
		if err != nil {
			continue
		}

		broadcast := d.directedBroadcast(inet)
		if broadcast != nil {
			addr := &net.UDPAddr{IP: broadcast, Port: MNDPPort}
			_, err = d.udp4Conn.WriteTo(packet, addr)
			if err != nil {
				fmt.Printf("Failed to send to %s: %v\n", broadcast, err)
			}
		}
	}

	return nil
}

// sendUDP6Multicast sends packet to IPv6 link-local all nodes
func (d *MNDPDiscoverer) sendUDP6Multicast(packet []byte) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range interfaces {
		if (iface.Flags&net.FlagUp != 0) && (iface.Flags&net.FlagMulticast != 0) {
			addr := &net.UDPAddr{
				IP:   net.ParseIP(IPv6LinkLocalAll),
				Port: MNDPPort,
				Zone: iface.Name,
			}
			_, err = d.udp6Conn.WriteTo(packet, addr)
			if err != nil {
				fmt.Printf("Failed to send to %s on %s: %v\n", IPv6LinkLocalAll, iface.Name, err)
			}
		}
	}

	return nil
}

// directedBroadcast calculates the directed broadcast address for a network
func (d *MNDPDiscoverer) directedBroadcast(inet *net.IPNet) net.IP {
	if inet.IP.To4() == nil {
		return nil
	}

	ip := make(net.IP, len(inet.IP.To4()))
	binary.BigEndian.PutUint32(ip,
		binary.BigEndian.Uint32(inet.IP.To4())|^binary.BigEndian.Uint32(net.IP(inet.Mask).To4()))
	return ip
}
