package goneighbors

import (
	"time"
)

// DiscoveryProtocol represents the protocol used to discover the device
type DiscoveryProtocol string

const (
	ProtocolMNDP DiscoveryProtocol = "MNDP"
	ProtocolCDP  DiscoveryProtocol = "CDP"
	ProtocolLLDP DiscoveryProtocol = "LLDP"
)

// Neighbor represents a discovered network device
type Neighbor struct {
	MACAddress     string
	Identity       string
	Version        string
	Platform       string
	Uptime         time.Duration
	SoftwareID     string
	Board          string
	UnpackedData   string
	IPv6Address    string
	InterfaceName  string
	IPAddress      string
	SSHPort        int
	HasSSH         bool
	HasTelnet      bool
	Protocol       DiscoveryProtocol
	DeviceID       string
	PortID         string
	SystemName     string
	SystemDesc     string
	PortDesc       string
	MgmtAddr       string
	TTL            uint16
	Capabilities   uint32
	NativeVLAN     uint16
	VLANs          []uint16
	PowerAvailable uint32
	PowerRequested uint32
	DiscoveredAt   time.Time
}

// MikroTikNeighbor represents a discovered MikroTik device (for backward compatibility)
type MikroTikNeighbor = Neighbor

// TLV represents a Type-Length-Value structure in MNDP packets
type TLV struct {
	Tag    TLVTag
	Length uint16
	Value  []byte
}

// TLVTag represents MNDP TLV tag types
type TLVTag uint16

const (
	TagMNDP          TLVTag = 0
	TagMACAddr       TLVTag = 1
	TagIdentity      TLVTag = 5
	TagVersion       TLVTag = 7
	TagPlatform      TLVTag = 8
	TagUptime        TLVTag = 10
	TagSoftwareID    TLVTag = 11
	TagBoard         TLVTag = 12
	TagUnpack        TLVTag = 14
	TagIPv6Addr      TLVTag = 15
	TagInterfaceName TLVTag = 16
	TagIPv4Addr      TLVTag = 17
	TagMAX           TLVTag = TagIPv4Addr
)

// MNDPPacket represents a MikroTik Neighbor Discovery Protocol packet
type MNDPPacket struct {
	TypeTag uint16
	SeqNo   uint16
	Fields  map[TLVTag]TLV
}

// DiscoveryResult contains the results of a neighbor discovery scan
type DiscoveryResult struct {
	TotalFound int
	WithSSH    int
	Neighbors  []Neighbor
	Duration   time.Duration
	ByProtocol map[DiscoveryProtocol]int
}
