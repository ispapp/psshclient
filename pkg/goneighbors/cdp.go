package goneighbors

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
	"unsafe"
)

const (
	// CDP uses SNAP encapsulation with these values
	CDPMulticast        = "01:00:0c:cc:cc:cc" // CDP multicast MAC address
	CDPEtherType        = 0x2000              // CDP EtherType (actually uses SNAP)
	CDPLLC              = 0xAAAA              // LLC header for CDP
	CDPOrganizationCode = 0x00000C            // Cisco organization code
	CDPProtocolID       = 0x2000              // CDP protocol ID in SNAP header
)

// CDP TLV Types (according to Wireshark documentation)
const (
	CDPTLVDeviceID           = 0x0001
	CDPTLVAddress            = 0x0002
	CDPTLVPortID             = 0x0003
	CDPTLVCapabilities       = 0x0004
	CDPTLVVersion            = 0x0005
	CDPTLVPlatform           = 0x0006
	CDPTLVIPPrefix           = 0x0007
	CDPTLVVTPDomain          = 0x0009
	CDPTLVNativeVLAN         = 0x000a
	CDPTLVDuplex             = 0x000b
	CDPTLVApplianceID        = 0x000c
	CDPTLVApplianceVLAN      = 0x000e
	CDPTLVPowerConsumption   = 0x0010
	CDPTLVMTU                = 0x0011
	CDPTLVExtendedTrust      = 0x0012
	CDPTLVUntrustedCOS       = 0x0013
	CDPTLVSystemName         = 0x0014
	CDPTLVSystemOID          = 0x0015
	CDPTLVMgmtAddress        = 0x0016
	CDPTLVLocation           = 0x0017
	CDPTLVExternalPortID     = 0x0018
	CDPTLVPowerRequested     = 0x0019
	CDPTLVPowerAvailable     = 0x001a
	CDPTLVPortUnidirectional = 0x001b
)

// CDP Capabilities
const (
	CDPCapRouter     = 0x01
	CDPCapTBBridge   = 0x02
	CDPCapSRBridge   = 0x04
	CDPCapSwitch     = 0x08
	CDPCapHost       = 0x10
	CDPCapIGMPFilter = 0x20
	CDPCapRepeater   = 0x40
	CDPCapPhone      = 0x80
	CDPCapRemote     = 0x100
)

// CDPHeader represents the CDP packet header
type CDPHeader struct {
	Version  uint8
	TTL      uint8
	Checksum uint16
}

// CDPTLV represents a CDP Type-Length-Value field
type CDPTLV struct {
	Type   uint16
	Length uint16
	Value  []byte
}

// CDPDiscoverer handles Cisco Discovery Protocol
// Note: This is a simulation since CDP requires raw Layer 2 access
type CDPDiscoverer struct {
	running  bool
	stopChan chan struct{}
}

// NewCDPDiscoverer creates a new CDP discoverer
func NewCDPDiscoverer() (*CDPDiscoverer, error) {
	return &CDPDiscoverer{
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins CDP discovery simulation
func (d *CDPDiscoverer) Start(neighborChan chan<- Neighbor) error {
	d.running = true

	go func() {
		fmt.Println("CDP Discovery started (requires raw socket access for real implementation)")

		// In a real implementation, this would:
		// 1. Create raw socket with AF_PACKET
		// 2. Bind to network interface
		// 3. Set packet filter for CDP multicast address
		// 4. Listen for CDP frames with SNAP encapsulation
		// 5. Parse CDP TLVs and extract neighbor information

		// For simulation, we'll just run a placeholder loop
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for d.running {
			select {
			case <-d.stopChan:
				return
			case <-ticker.C:
				// Placeholder: In real implementation, this would send CDP advertisements
				// and process received CDP packets
				fmt.Println("CDP: Simulating neighbor discovery (requires root privileges for real implementation)")
			}
		}
	}()

	return nil
}

// Stop stops the CDP discoverer
func (d *CDPDiscoverer) Stop() {
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

// InjectCDPNeighbor allows manual injection of CDP neighbor data for testing
func (d *CDPDiscoverer) InjectCDPNeighbor(deviceID, platform, version, portID, mgmtAddr string, capabilities uint32, neighborChan chan<- Neighbor) {
	neighbor := Neighbor{
		DeviceID:     deviceID,
		Platform:     platform,
		Version:      version,
		PortID:       portID,
		IPAddress:    mgmtAddr,
		MgmtAddr:     mgmtAddr,
		Capabilities: capabilities,
		Protocol:     ProtocolCDP,
		DiscoveredAt: time.Now(),
		HasSSH:       true, // Assume Cisco devices have SSH
		HasTelnet:    true, // Assume Cisco devices have Telnet
		SSHPort:      22,
	}

	if neighborChan != nil {
		select {
		case neighborChan <- neighbor:
		default:
			// Channel full, skip
		}
	}
}

// Note: The following methods would be used in a real implementation with raw sockets

// parseCDPFrame would parse a raw CDP frame from Layer 2
// It would:
// 1. Verify Ethernet header with CDP multicast MAC
// 2. Parse LLC/SNAP headers
// 3. Extract CDP header (version, TTL, checksum)
// 4. Parse CDP TLVs
// 5. Return structured neighbor information

// sendCDPAdvertisement would send CDP advertisements
// It would:
// 1. Build CDP frame with proper SNAP encapsulation
// 2. Add required TLVs (Device ID, Platform, etc.)
// 3. Calculate checksum
// 4. Send via raw socket to CDP multicast address

// For a complete implementation, you would need:
// - Root privileges for raw socket access
// - Proper error handling for network operations
// - Platform-specific network interface binding
// - CDP frame validation and security checks

// ParseCDPData is a public method that can be used for testing or manual CDP packet processing
// It wraps the internal parseCDPPacket method for external use
func (d *CDPDiscoverer) ParseCDPData(data []byte) (*Neighbor, error) {
	return d.parseCDPPacket(data)
}

// parseCDPPacket parses a CDP packet and returns a Neighbor
// This method is ready for real CDP implementation when raw socket access is added
func (d *CDPDiscoverer) parseCDPPacket(data []byte) (*Neighbor, error) {
	if len(data) < 14 {
		return nil, fmt.Errorf("packet too short for Ethernet header")
	}

	// Check if it's a CDP packet
	etherType := binary.BigEndian.Uint16(data[12:14])
	if etherType != CDPEtherType {
		return nil, fmt.Errorf("not a CDP packet")
	}

	// Skip Ethernet header (14 bytes) and SNAP header (8 bytes)
	if len(data) < 22 {
		return nil, fmt.Errorf("packet too short for CDP")
	}

	cdpData := data[22:]
	if len(cdpData) < 4 {
		return nil, fmt.Errorf("CDP data too short")
	}

	// CDP header: Version (1) + TTL (1) + Checksum (2)
	version := cdpData[0]
	ttl := cdpData[1]
	// checksum := binary.BigEndian.Uint16(cdpData[2:4])

	if version != 1 && version != 2 {
		return nil, fmt.Errorf("unsupported CDP version: %d", version)
	}

	neighbor := &Neighbor{
		Protocol:     ProtocolCDP,
		TTL:          uint16(ttl),
		DiscoveredAt: time.Now(),
		SSHPort:      22,
		HasSSH:       true, // Assume Cisco devices have SSH
		HasTelnet:    true, // Assume Cisco devices have Telnet
	}

	// Parse TLVs starting from offset 4
	offset := 4
	for offset < len(cdpData) {
		if offset+4 > len(cdpData) {
			break
		}

		tlvType := binary.BigEndian.Uint16(cdpData[offset : offset+2])
		tlvLength := binary.BigEndian.Uint16(cdpData[offset+2 : offset+4])

		if tlvLength < 4 || offset+int(tlvLength) > len(cdpData) {
			break
		}

		tlvValue := cdpData[offset+4 : offset+int(tlvLength)]
		d.processCDPTLV(tlvType, tlvValue, neighbor)

		offset += int(tlvLength)
	}

	return neighbor, nil
}

// processCDPTLV processes a CDP TLV and updates the neighbor information
// This method is ready for real CDP implementation when raw socket access is added
func (d *CDPDiscoverer) processCDPTLV(tlvType uint16, tlvValue []byte, neighbor *Neighbor) {
	switch tlvType {
	case CDPTLVDeviceID:
		neighbor.DeviceID = strings.TrimSpace(string(tlvValue))
		neighbor.Identity = neighbor.DeviceID
	case CDPTLVAddress:
		// Parse address TLV (complex structure)
		if len(tlvValue) >= 8 {
			numAddresses := binary.BigEndian.Uint32(tlvValue[0:4])
			if numAddresses > 0 && len(tlvValue) >= 13 {
				// Skip to first address (simplified parsing)
				addrType := tlvValue[5]
				addrLen := tlvValue[6]
				if addrType == 1 && addrLen == 4 && len(tlvValue) >= 7+int(addrLen) {
					ip := net.IPv4(tlvValue[7], tlvValue[8], tlvValue[9], tlvValue[10])
					neighbor.IPAddress = ip.String()
				}
			}
		}
	case CDPTLVPortID:
		neighbor.PortID = strings.TrimSpace(string(tlvValue))
	case CDPTLVCapabilities:
		if len(tlvValue) >= 4 {
			neighbor.Capabilities = binary.BigEndian.Uint32(tlvValue)
		}
	case CDPTLVVersion:
		neighbor.Version = strings.TrimSpace(string(tlvValue))
	case CDPTLVPlatform:
		neighbor.Platform = strings.TrimSpace(string(tlvValue))
	case CDPTLVNativeVLAN:
		if len(tlvValue) >= 2 {
			neighbor.NativeVLAN = binary.BigEndian.Uint16(tlvValue)
		}
	case CDPTLVPowerRequested:
		if len(tlvValue) >= 2 {
			neighbor.PowerRequested = uint32(binary.BigEndian.Uint16(tlvValue))
		}
	case CDPTLVSystemName:
		neighbor.SystemName = strings.TrimSpace(string(tlvValue))
	case CDPTLVMgmtAddress:
		if len(tlvValue) >= 4 {
			ip := net.IPv4(tlvValue[0], tlvValue[1], tlvValue[2], tlvValue[3])
			neighbor.MgmtAddr = ip.String()
			if neighbor.IPAddress == "" {
				neighbor.IPAddress = neighbor.MgmtAddr
			}
		}
	}
}

// SendCDPPacket sends a CDP advertisement (requires root privileges)
func (d *CDPDiscoverer) SendCDPPacket(deviceID, platform string) error {
	// This is a simplified CDP packet creation
	// In production, you'd need proper interface binding and packet construction
	return fmt.Errorf("CDP packet sending not implemented (requires root privileges)")
}

// htons converts host byte order to network byte order
func htons(i uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return *(*uint16)(unsafe.Pointer(&b[0]))
}
