package goneighbors

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// LLDP Constants based on IEEE 802.1AB
const (
	LLDPMulticast   = "01:80:C2:00:00:0E" // LLDP multicast MAC address
	LLDPEtherType   = 0x88CC              // LLDP EtherType
	LLDPDefaultTTL  = 120                 // Default TTL in seconds
	LLDPTxInterval  = 30                  // Transmission interval in seconds
	LLDPTxHoldMulti = 4                   // Hold multiplier
	LLDPReinitDelay = 2                   // Reinit delay in seconds
	LLDPTxDelay     = 2                   // Tx delay in seconds
)

// LLDP TLV Types
const (
	LLDPTLVEndOfLLDPDU     = 0   // Mandatory
	LLDPTLVChassisID       = 1   // Mandatory
	LLDPTLVPortID          = 2   // Mandatory
	LLDPTLVTimeToLive      = 3   // Mandatory
	LLDPTLVPortDescription = 4   // Optional
	LLDPTLVSystemName      = 5   // Optional
	LLDPTLVSystemDesc      = 6   // Optional
	LLDPTLVSystemCaps      = 7   // Optional
	LLDPTLVMgmtAddr        = 8   // Optional
	LLDPTLVOrgSpecific     = 127 // Organizationally Specific TLVs
)

// LLDP Chassis ID Subtypes
const (
	LLDPChassisIDChassisComponent = 1
	LLDPChassisIDInterfaceAlias   = 2
	LLDPChassisIDPortComponent    = 3
	LLDPChassisIDMACAddress       = 4
	LLDPChassisIDNetworkAddress   = 5
	LLDPChassisIDInterfaceName    = 6
	LLDPChassisIDLocallyAssigned  = 7
)

// LLDP Port ID Subtypes
const (
	LLDPPortIDInterfaceAlias  = 1
	LLDPPortIDPortComponent   = 2
	LLDPPortIDMACAddress      = 3
	LLDPPortIDNetworkAddress  = 4
	LLDPPortIDInterfaceName   = 5
	LLDPPortIDAgentCircuitID  = 6
	LLDPPortIDLocallyAssigned = 7
)

// LLDP System Capabilities
const (
	LLDPSysCapOther       = 0x01
	LLDPSysCapRepeater    = 0x02
	LLDPSysCapBridge      = 0x04
	LLDPSysCapWLANAP      = 0x08
	LLDPSysCapRouter      = 0x10
	LLDPSysCapTelephone   = 0x20
	LLDPSysCapDocsisCable = 0x40
	LLDPSysCapStationOnly = 0x80
)

// LLDPTLV represents an LLDP Type-Length-Value field
type LLDPTLV struct {
	Type   uint8
	Length uint16
	Value  []byte
}

// LLDPFrame represents a complete LLDP frame
type LLDPFrame struct {
	ChassisID       *LLDPTLV
	PortID          *LLDPTLV
	TTL             *LLDPTLV
	PortDescription *LLDPTLV
	SystemName      *LLDPTLV
	SystemDesc      *LLDPTLV
	SystemCaps      *LLDPTLV
	MgmtAddr        *LLDPTLV
	OrgTLVs         []*LLDPTLV
}

// LLDPDiscoverer handles Link Layer Discovery Protocol
type LLDPDiscoverer struct {
	running    bool
	stopChan   chan struct{}
	interfaces []net.Interface
	packetMode bool // true for raw packet mode, false for system command mode
}

// NewLLDPDiscoverer creates a new LLDP discoverer
func NewLLDPDiscoverer() (*LLDPDiscoverer, error) {
	discoverer := &LLDPDiscoverer{
		stopChan: make(chan struct{}),
	}

	// Try to determine available interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	discoverer.interfaces = interfaces

	// For now, default to system command mode since raw packets require privileges
	// In the future, we can detect capabilities and switch modes
	discoverer.packetMode = false

	return discoverer, nil
}

// Start begins LLDP discovery
func (d *LLDPDiscoverer) Start(neighborChan chan<- Neighbor) error {
	d.running = true

	go func() {
		fmt.Println("LLDP Discovery started")

		ticker := time.NewTicker(LLDPTxInterval * time.Second)
		defer ticker.Stop()

		// Immediate discovery
		d.performDiscovery(neighborChan)

		for d.running {
			select {
			case <-d.stopChan:
				return
			case <-ticker.C:
				if d.running {
					d.performDiscovery(neighborChan)
				}
			}
		}
	}()

	return nil
}

// Stop stops the LLDP discoverer
func (d *LLDPDiscoverer) Stop() {
	if !d.running {
		return // Already stopped
	}

	d.running = false

	if d.stopChan != nil {
		// Safely close the channel only if it's not already closed
		select {
		case <-d.stopChan:
			// Channel is already closed
		default:
			close(d.stopChan)
		}
	}
}

// performDiscovery performs the actual LLDP discovery
func (d *LLDPDiscoverer) performDiscovery(neighborChan chan<- Neighbor) {
	if d.packetMode {
		d.performRawPacketDiscovery(neighborChan)
	} else {
		d.performSystemCommandDiscovery(neighborChan)
	}
}

// performRawPacketDiscovery performs LLDP discovery using raw packets
func (d *LLDPDiscoverer) performRawPacketDiscovery(neighborChan chan<- Neighbor) {
	// Check if we have sufficient privileges for raw sockets
	if os.Geteuid() != 0 {
		fmt.Println("Raw packet LLDP discovery requires root privileges, falling back to system commands")
		d.performSystemCommandDiscovery(neighborChan)
		return
	}

	fmt.Println("Starting raw packet LLDP discovery...")

	// Iterate through network interfaces and send LLDP advertisements
	for _, iface := range d.interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 || iface.HardwareAddr == nil {
			continue
		}

		// Build and potentially send LLDP packet (simulated for now)
		packet := d.buildLLDPPacket(iface.HardwareAddr, iface.Name)
		if len(packet) > 0 {
			fmt.Printf("Built LLDP packet for interface %s (MAC: %s): %d bytes\n",
				iface.Name, iface.HardwareAddr.String(), len(packet))
		}

		// For now, fall back to other discovery methods
		// In a real implementation, this would:
		// 1. Create raw socket with appropriate platform-specific code
		// 2. Send the LLDP packet
		// 3. Listen for responses
		go d.simulateLLDPDiscovery(iface, neighborChan)
	}

	// If no interfaces were processed, fall back to system commands
	d.performSystemCommandDiscovery(neighborChan)
}

// simulateLLDPDiscovery simulates LLDP discovery for development/testing
func (d *LLDPDiscoverer) simulateLLDPDiscovery(iface net.Interface, neighborChan chan<- Neighbor) {
	// This is a placeholder that simulates finding LLDP neighbors
	// In production, this would be replaced with actual raw socket implementation

	time.Sleep(1 * time.Second) // Simulate network delay

	// Try to discover potential neighbors on this interface's network
	addrs, err := iface.Addrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			// Simulate discovering a network device
			d.scanSubnetForLLDP(ipnet, neighborChan)
			break
		}
	}
} // buildLLDPPacket builds an LLDP packet for transmission (future implementation)
func (d *LLDPDiscoverer) buildLLDPPacket(srcMAC net.HardwareAddr, ifaceName string) []byte {
	// This would implement the packet building logic from your guide
	// Building Ethernet frame with LLDP payload
	var packet []byte

	// Ethernet header (14 bytes)
	dstMAC, _ := net.ParseMAC(LLDPMulticast)
	packet = append(packet, dstMAC...) // Destination MAC (6 bytes)
	packet = append(packet, srcMAC...) // Source MAC (6 bytes)
	etherType := make([]byte, 2)
	binary.BigEndian.PutUint16(etherType, LLDPEtherType)
	packet = append(packet, etherType...) // EtherType (2 bytes)

	// LLDP TLVs
	packet = append(packet, d.buildChassisIDTLV(srcMAC)...)
	packet = append(packet, d.buildPortIDTLV(ifaceName)...)
	packet = append(packet, d.buildTTLTLV(LLDPDefaultTTL)...)
	packet = append(packet, d.buildSystemNameTLV("Go-LLDP-Device")...)
	packet = append(packet, d.buildEndOfLLDPDUTLV()...)

	return packet
}

// buildChassisIDTLV builds a Chassis ID TLV
func (d *LLDPDiscoverer) buildChassisIDTLV(mac net.HardwareAddr) []byte {
	tlv := make([]byte, 0)

	// TLV Header: Type (7 bits) + Length (9 bits) = 2 bytes
	length := uint16(1 + len(mac)) // Subtype (1 byte) + MAC (6 bytes)
	header := (uint16(LLDPTLVChassisID) << 9) | length
	headerBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(headerBytes, header)
	tlv = append(tlv, headerBytes...)

	// Subtype + Value
	tlv = append(tlv, LLDPChassisIDMACAddress) // Subtype
	tlv = append(tlv, mac...)                  // MAC address

	return tlv
}

// buildPortIDTLV builds a Port ID TLV
func (d *LLDPDiscoverer) buildPortIDTLV(ifaceName string) []byte {
	tlv := make([]byte, 0)

	length := uint16(1 + len(ifaceName)) // Subtype + interface name
	header := (uint16(LLDPTLVPortID) << 9) | length
	headerBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(headerBytes, header)
	tlv = append(tlv, headerBytes...)

	tlv = append(tlv, LLDPPortIDInterfaceName) // Subtype
	tlv = append(tlv, []byte(ifaceName)...)    // Interface name

	return tlv
}

// buildTTLTLV builds a TTL TLV
func (d *LLDPDiscoverer) buildTTLTLV(ttl uint16) []byte {
	tlv := make([]byte, 0)

	length := uint16(2) // TTL is 2 bytes
	header := (uint16(LLDPTLVTimeToLive) << 9) | length
	headerBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(headerBytes, header)
	tlv = append(tlv, headerBytes...)

	ttlBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(ttlBytes, ttl)
	tlv = append(tlv, ttlBytes...)

	return tlv
}

// buildSystemNameTLV builds a System Name TLV
func (d *LLDPDiscoverer) buildSystemNameTLV(name string) []byte {
	tlv := make([]byte, 0)

	length := uint16(len(name))
	header := (uint16(LLDPTLVSystemName) << 9) | length
	headerBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(headerBytes, header)
	tlv = append(tlv, headerBytes...)

	tlv = append(tlv, []byte(name)...)

	return tlv
}

// buildEndOfLLDPDUTLV builds an End of LLDPDU TLV
func (d *LLDPDiscoverer) buildEndOfLLDPDUTLV() []byte {
	// End of LLDPDU is always Type=0, Length=0
	return []byte{0x00, 0x00}
}

// decodeLLDPFrame decodes an LLDP frame from raw bytes
func (d *LLDPDiscoverer) decodeLLDPFrame(frameData []byte) (*LLDPFrame, error) {
	if len(frameData) < 14 {
		return nil, fmt.Errorf("frame too short for Ethernet header")
	}

	// Skip Ethernet header (14 bytes)
	lldpData := frameData[14:]
	frame := &LLDPFrame{}

	offset := 0
	for offset < len(lldpData) {
		if offset+2 > len(lldpData) {
			break
		}

		// Parse TLV header
		header := binary.BigEndian.Uint16(lldpData[offset : offset+2])
		tlvType := uint8(header >> 9)
		tlvLength := header & 0x1FF

		if offset+2+int(tlvLength) > len(lldpData) {
			return nil, fmt.Errorf("invalid TLV length")
		}

		tlvValue := lldpData[offset+2 : offset+2+int(tlvLength)]

		tlv := &LLDPTLV{
			Type:   tlvType,
			Length: tlvLength,
			Value:  tlvValue,
		}

		// Store TLV in appropriate field
		switch tlvType {
		case LLDPTLVEndOfLLDPDU:
			return frame, nil // End of frame
		case LLDPTLVChassisID:
			frame.ChassisID = tlv
		case LLDPTLVPortID:
			frame.PortID = tlv
		case LLDPTLVTimeToLive:
			frame.TTL = tlv
		case LLDPTLVPortDescription:
			frame.PortDescription = tlv
		case LLDPTLVSystemName:
			frame.SystemName = tlv
		case LLDPTLVSystemDesc:
			frame.SystemDesc = tlv
		case LLDPTLVSystemCaps:
			frame.SystemCaps = tlv
		case LLDPTLVMgmtAddr:
			frame.MgmtAddr = tlv
		case LLDPTLVOrgSpecific:
			frame.OrgTLVs = append(frame.OrgTLVs, tlv)
		}

		offset += 2 + int(tlvLength)
	}

	return frame, nil
}

// frameToNeighbor converts an LLDP frame to a Neighbor struct
func (d *LLDPDiscoverer) frameToNeighbor(frame *LLDPFrame) Neighbor {
	neighbor := Neighbor{
		Protocol:     ProtocolLLDP,
		DiscoveredAt: time.Now(),
		HasSSH:       true, // Assume managed devices have SSH
		HasTelnet:    true, // Assume managed devices have Telnet
		SSHPort:      22,
	}

	// Extract Chassis ID
	if frame.ChassisID != nil && len(frame.ChassisID.Value) > 1 {
		subtype := frame.ChassisID.Value[0]
		if subtype == LLDPChassisIDMACAddress && len(frame.ChassisID.Value) >= 7 {
			mac := net.HardwareAddr(frame.ChassisID.Value[1:7])
			neighbor.DeviceID = mac.String()
		} else {
			neighbor.DeviceID = string(frame.ChassisID.Value[1:])
		}
	}

	// Extract Port ID
	if frame.PortID != nil && len(frame.PortID.Value) > 1 {
		neighbor.PortID = string(frame.PortID.Value[1:])
	}

	// Extract System Name
	if frame.SystemName != nil {
		neighbor.SystemName = string(frame.SystemName.Value)
		neighbor.Identity = neighbor.SystemName
	}

	// Extract System Description
	if frame.SystemDesc != nil {
		neighbor.SystemDesc = string(frame.SystemDesc.Value)
		d.extractPlatformFromDescription(&neighbor)
	}

	// Extract Management Address
	if frame.MgmtAddr != nil {
		neighbor.MgmtAddr = d.ExtractMgmtAddress(frame.MgmtAddr.Value)
		neighbor.IPAddress = neighbor.MgmtAddr
	}

	// Extract Port Description
	if frame.PortDescription != nil {
		neighbor.PortDesc = string(frame.PortDescription.Value)
	}

	return neighbor
}

// extractMgmtAddress extracts management address from LLDP management address TLV
func (d *LLDPDiscoverer) ExtractMgmtAddress(tlvValue []byte) string {
	if len(tlvValue) < 2 {
		return ""
	}

	addrLen := tlvValue[0]
	if len(tlvValue) < int(1+addrLen) {
		return ""
	}

	addrSubtype := tlvValue[1]
	addrBytes := tlvValue[2 : 1+addrLen]

	switch addrSubtype {
	case 1: // IPv4
		if len(addrBytes) == 4 {
			return net.IP(addrBytes).String()
		}
	case 2: // IPv6
		if len(addrBytes) == 16 {
			return net.IP(addrBytes).String()
		}
	}

	return ""
}

// performSystemCommandDiscovery performs LLDP discovery using system commands (fallback)
func (d *LLDPDiscoverer) performSystemCommandDiscovery(neighborChan chan<- Neighbor) {
	// Try different LLDP commands based on the system
	commands := [][]string{
		{"lldpctl"},                      // lldpd
		{"lldpcli", "show", "neighbors"}, // lldpd alternative
		{"lldptool"},                     // lldpad
	}

	for _, cmd := range commands {
		if neighbors := d.tryLLDPCommand(cmd); len(neighbors) > 0 {
			for _, neighbor := range neighbors {
				if d.running {
					neighborChan <- neighbor
				}
			}
			return // Successfully found neighbors with this command
		}
	}

	// If no system LLDP tools are available, try network scanning for LLDP-capable devices
	d.scanForLLDPDevices(neighborChan)
}

// tryLLDPCommand tries to execute an LLDP command and parse its output
func (d *LLDPDiscoverer) tryLLDPCommand(cmdArgs []string) []Neighbor {
	var neighbors []Neighbor

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return neighbors
	}

	// Parse the output based on the command
	if cmdArgs[0] == "lldpctl" || (len(cmdArgs) > 1 && cmdArgs[1] == "show") {
		neighbors = d.parseLLDPCtlOutput(string(output))
	} else if cmdArgs[0] == "lldptool" {
		neighbors = d.parseLLDPToolOutput(string(output))
	}

	return neighbors
}

// parseLLDPCtlOutput parses lldpctl command output
func (d *LLDPDiscoverer) parseLLDPCtlOutput(output string) []Neighbor {
	var neighbors []Neighbor

	lines := strings.Split(output, "\n")
	var currentNeighbor *Neighbor

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for neighbor entries
		if strings.Contains(line, "Interface:") {
			if currentNeighbor != nil {
				neighbors = append(neighbors, *currentNeighbor)
			}
			currentNeighbor = &Neighbor{
				Protocol:     ProtocolLLDP,
				DiscoveredAt: time.Now(),
				SSHPort:      22,
				HasSSH:       true,
				HasTelnet:    true,
			}

			// Extract interface name
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				currentNeighbor.InterfaceName = strings.TrimSpace(parts[1])
			}
		}

		if currentNeighbor == nil {
			continue
		}

		// Parse various LLDP fields
		if strings.Contains(line, "ChassisID:") {
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				currentNeighbor.DeviceID = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "SysName:") {
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				currentNeighbor.SystemName = strings.TrimSpace(parts[1])
				currentNeighbor.Identity = currentNeighbor.SystemName
			}
		} else if strings.Contains(line, "SysDescr:") {
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				currentNeighbor.SystemDesc = strings.TrimSpace(parts[1])
				// Try to extract platform from system description
				d.extractPlatformFromDescription(currentNeighbor)
			}
		} else if strings.Contains(line, "PortID:") {
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				currentNeighbor.PortID = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "PortDescr:") {
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				currentNeighbor.PortDesc = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "MgmtIP:") || strings.Contains(line, "ManagementIP:") {
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				ipStr := strings.TrimSpace(parts[1])
				if ip := net.ParseIP(ipStr); ip != nil {
					currentNeighbor.IPAddress = ip.String()
					currentNeighbor.MgmtAddr = ip.String()
				}
			}
		}
	}

	// Add the last neighbor
	if currentNeighbor != nil {
		neighbors = append(neighbors, *currentNeighbor)
	}

	return neighbors
}

// parseLLDPToolOutput parses lldptool command output
func (d *LLDPDiscoverer) parseLLDPToolOutput(output string) []Neighbor {
	var neighbors []Neighbor

	// lldptool output parsing would be implemented here
	// This is a simplified version
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Neighbor") {
			neighbor := Neighbor{
				Protocol:     ProtocolLLDP,
				DiscoveredAt: time.Now(),
				SSHPort:      22,
				HasSSH:       true,
				HasTelnet:    true,
			}

			// Extract basic information from the line
			if parts := strings.Fields(line); len(parts) > 1 {
				neighbor.Identity = parts[1]
			}

			neighbors = append(neighbors, neighbor)
		}
	}

	return neighbors
}

// scanForLLDPDevices scans the network for LLDP-capable devices
func (d *LLDPDiscoverer) scanForLLDPDevices(neighborChan chan<- Neighbor) {
	// Get local network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					// Scan the local subnet for potential LLDP devices
					d.scanSubnetForLLDP(ipnet, neighborChan)
				}
			}
		}
	}
}

// scanSubnetForLLDP scans a subnet for devices that might support LLDP
func (d *LLDPDiscoverer) scanSubnetForLLDP(ipnet *net.IPNet, neighborChan chan<- Neighbor) {
	// This is a simplified network scan
	// In practice, you'd need raw sockets to detect LLDP frames

	network := ipnet.IP.Mask(ipnet.Mask)
	broadcast := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		broadcast[i] = network[i] | ^ipnet.Mask[i]
	}

	// Scan a few IPs in the subnet to look for managed switches/routers
	// that typically support LLDP
	commonIPs := []string{
		fmt.Sprintf("%d.%d.%d.1", network[0], network[1], network[2]),
		fmt.Sprintf("%d.%d.%d.254", network[0], network[1], network[2]),
		fmt.Sprintf("%d.%d.%d.10", network[0], network[1], network[2]),
		fmt.Sprintf("%d.%d.%d.20", network[0], network[1], network[2]),
	}

	for _, ip := range commonIPs {
		if !d.running {
			break
		}

		if d.probeLLDPDevice(ip, neighborChan) {
			// Found a potential LLDP device
		}
	}
}

// probeLLDPDevice probes an IP to see if it might be an LLDP-capable device
func (d *LLDPDiscoverer) probeLLDPDevice(ip string, neighborChan chan<- Neighbor) bool {
	// Try to connect to common management ports
	managementPorts := []int{80, 443, 23, 22, 161} // HTTP, HTTPS, Telnet, SSH, SNMP

	for _, port := range managementPorts {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 2*time.Second)
		if err == nil {
			conn.Close()

			// If we can connect, this might be a managed device
			neighbor := Neighbor{
				IPAddress:    ip,
				Protocol:     ProtocolLLDP,
				DiscoveredAt: time.Now(),
				Platform:     "Unknown",
				Identity:     ip,
				SSHPort:      22,
				HasSSH:       port == 22,
				HasTelnet:    port == 23,
			}

			// Try to get more information via reverse DNS
			if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
				neighbor.SystemName = names[0]
				neighbor.Identity = names[0]
			}

			neighborChan <- neighbor
			return true
		}
	}

	return false
}

// extractPlatformFromDescription tries to extract platform information from system description
func (d *LLDPDiscoverer) extractPlatformFromDescription(neighbor *Neighbor) {
	desc := strings.ToLower(neighbor.SystemDesc)

	// Define platform patterns
	platformPatterns := map[string]*regexp.Regexp{
		"Cisco":    regexp.MustCompile(`cisco|catalyst|nexus`),
		"MikroTik": regexp.MustCompile(`mikrotik|routeros`),
		"Juniper":  regexp.MustCompile(`juniper|junos`),
		"HPE":      regexp.MustCompile(`hewlett.*packard|procurve|aruba`),
		"Dell":     regexp.MustCompile(`dell|powerconnect`),
		"Ubiquiti": regexp.MustCompile(`ubiquiti|unifi|edgemax`),
		"Netgear":  regexp.MustCompile(`netgear`),
		"D-Link":   regexp.MustCompile(`d-link`),
		"TP-Link":  regexp.MustCompile(`tp-link`),
	}

	for platform, pattern := range platformPatterns {
		if pattern.MatchString(desc) {
			neighbor.Platform = platform
			break
		}
	}

	// If no platform detected, set to "Unknown"
	if neighbor.Platform == "" {
		neighbor.Platform = "Unknown"
	}
}
