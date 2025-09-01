package goneighbors

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// NeighborScanner handles the overall neighbor discovery process
type NeighborScanner struct {
	mndp       *MNDPDiscoverer
	cdp        *CDPDiscoverer
	lldp       *LLDPDiscoverer
	neighbors  map[string]Neighbor
	mu         sync.RWMutex
	updateChan chan Neighbor
	running    bool
}

// NewNeighborScanner creates a new neighbor scanner
func NewNeighborScanner() *NeighborScanner {
	return &NeighborScanner{
		neighbors:  make(map[string]Neighbor),
		updateChan: make(chan Neighbor, 100),
	}
}

// StartDiscoveryWithoutSSHCheck starts the neighbor discovery process without SSH connectivity checks
func (ns *NeighborScanner) StartDiscoveryWithoutSSHCheck(duration time.Duration) (*DiscoveryResult, error) {
	startTime := time.Now()

	// Initialize all discovery protocols
	var err error
	ns.mndp, err = NewMNDPDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("failed to create MNDP discoverer: %v", err)
	}

	ns.cdp, err = NewCDPDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("failed to create CDP discoverer: %v", err)
	}

	ns.lldp, err = NewLLDPDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("failed to create LLDP discoverer: %v", err)
	}

	// Create channels for each protocol
	mndpChan := make(chan Neighbor, 50)
	cdpChan := make(chan Neighbor, 50)
	lldpChan := make(chan Neighbor, 50)

	// Start all discoverers
	if err := ns.mndp.Start(mndpChan); err != nil {
		return nil, fmt.Errorf("failed to start MNDP discoverer: %v", err)
	}

	if err := ns.cdp.Start(cdpChan); err != nil {
		return nil, fmt.Errorf("failed to start CDP discoverer: %v", err)
	}

	if err := ns.lldp.Start(lldpChan); err != nil {
		return nil, fmt.Errorf("failed to start LLDP discoverer: %v", err)
	}

	ns.running = true

	// Send discovery packets periodically (MNDP only for now)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for ns.running {
			<-ticker.C
			if ns.running {
				ns.mndp.SendMNDPDiscovery()
			}
		}
	}()

	// Process discovered neighbors from all protocols
	go func() {
		for ns.running {
			select {
			case neighbor := <-mndpChan:
				ns.processNeighbor(neighbor)
			case neighbor := <-cdpChan:
				ns.processNeighbor(neighbor)
			case neighbor := <-lldpChan:
				ns.processNeighbor(neighbor)
			case <-time.After(100 * time.Millisecond):
				// Continue loop
			}
		}
	}()

	// Wait for the specified duration
	time.Sleep(duration)

	// Stop discovery
	ns.Stop()

	// Prepare results without SSH checking
	neighbors := ns.GetNeighbors()
	result := &DiscoveryResult{
		TotalFound: len(neighbors),
		Neighbors:  neighbors,
		Duration:   time.Since(startTime),
		WithSSH:    0, // Not checking SSH, so set to 0
		ByProtocol: make(map[DiscoveryProtocol]int),
	}

	// Count neighbors by protocol
	for _, neighbor := range neighbors {
		result.ByProtocol[neighbor.Protocol]++
	}

	return result, nil
}

// StartDiscovery starts the neighbor discovery process
func (ns *NeighborScanner) StartDiscovery(duration time.Duration) (*DiscoveryResult, error) {
	startTime := time.Now()

	var err error
	ns.mndp, err = NewMNDPDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("failed to create MNDP discoverer: %v", err)
	}

	mndpChan := make(chan MikroTikNeighbor, 50)
	if err := ns.mndp.Start(mndpChan); err != nil {
		return nil, fmt.Errorf("failed to start MNDP discoverer: %v", err)
	}

	ns.running = true

	// Send discovery packets periodically
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for ns.running {
			<-ticker.C
			if ns.running {
				ns.mndp.SendMNDPDiscovery()
			}
		}
	}()

	// Process discovered neighbors
	go func() {
		for neighbor := range mndpChan {
			if !ns.running {
				break
			}
			ns.processNeighbor(neighbor)
		}
	}()

	// Wait for the specified duration
	time.Sleep(duration)

	// Stop discovery
	ns.Stop()

	// Check SSH connectivity for all discovered neighbors
	ns.checkSSHConnectivity()

	// Prepare results
	neighbors := ns.GetNeighbors()
	result := &DiscoveryResult{
		TotalFound: len(neighbors),
		Neighbors:  neighbors,
		Duration:   time.Since(startTime),
	}

	// Count neighbors with SSH
	for _, neighbor := range neighbors {
		if neighbor.HasSSH {
			result.WithSSH++
		}
	}

	return result, nil
}

// Stop stops the neighbor scanner
func (ns *NeighborScanner) Stop() {
	if !ns.running {
		return // Already stopped
	}

	ns.running = false

	if ns.mndp != nil {
		ns.mndp.Stop()
	}
	if ns.cdp != nil {
		ns.cdp.Stop()
	}
	if ns.lldp != nil {
		ns.lldp.Stop()
	}
}

// processNeighbor processes a discovered neighbor from any protocol
func (ns *NeighborScanner) processNeighbor(neighbor Neighbor) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	key := neighbor.IPAddress
	if key == "" {
		key = neighbor.MACAddress
	}
	if key == "" {
		key = neighbor.DeviceID
	}

	// Update existing neighbor or add new one
	if existing, exists := ns.neighbors[key]; exists {
		// Merge information, keeping the most complete data
		if neighbor.Identity != "" {
			existing.Identity = neighbor.Identity
		}
		if neighbor.Platform != "" {
			existing.Platform = neighbor.Platform
		}
		if neighbor.Version != "" {
			existing.Version = neighbor.Version
		}
		if neighbor.IPAddress != "" {
			existing.IPAddress = neighbor.IPAddress
		}
		if neighbor.SystemName != "" {
			existing.SystemName = neighbor.SystemName
		}
		if neighbor.SystemDesc != "" {
			existing.SystemDesc = neighbor.SystemDesc
		}
		// Update discovery time to the most recent
		if neighbor.DiscoveredAt.After(existing.DiscoveredAt) {
			existing.DiscoveredAt = neighbor.DiscoveredAt
		}
		ns.neighbors[key] = existing
	} else {
		ns.neighbors[key] = neighbor
	}

	// Send to update channel
	select {
	case ns.updateChan <- neighbor:
	default:
		// Channel full, skip
	}

	fmt.Printf("Discovered %s neighbor: %s (%s) - %s\n",
		neighbor.Protocol, neighbor.Identity, neighbor.IPAddress, neighbor.Platform)
}

// GetNeighbors returns all discovered neighbors
func (ns *NeighborScanner) GetNeighbors() []Neighbor {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	neighbors := make([]Neighbor, 0, len(ns.neighbors))
	for _, neighbor := range ns.neighbors {
		neighbors = append(neighbors, neighbor)
	}
	return neighbors
}

// GetUpdateChannel returns the update channel
func (ns *NeighborScanner) GetUpdateChannel() <-chan Neighbor {
	return ns.updateChan
}

// checkSSHConnectivity checks SSH connectivity for discovered neighbors
func (ns *NeighborScanner) checkSSHConnectivity() {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	for key, neighbor := range ns.neighbors {
		if neighbor.IPAddress == "" {
			continue
		}

		// Check common SSH ports
		sshPorts := []int{22, 2222, 222}
		neighbor.HasSSH = false

		for _, port := range sshPorts {
			if ns.checkPortConnectivity(neighbor.IPAddress, port) {
				neighbor.HasSSH = true
				neighbor.SSHPort = port
				break
			}
		}

		// Check Telnet port 23
		neighbor.HasTelnet = ns.checkPortConnectivity(neighbor.IPAddress, 23)

		ns.neighbors[key] = neighbor
	}
}

// checkPortConnectivity checks if a port is open on the given IP
func (ns *NeighborScanner) checkPortConnectivity(ip string, port int) bool {
	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetNeighborsWithSSH returns only neighbors that have SSH connectivity
func (ns *NeighborScanner) GetNeighborsWithSSH() []Neighbor {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	var sshNeighbors []Neighbor
	for _, neighbor := range ns.neighbors {
		if neighbor.HasSSH {
			sshNeighbors = append(sshNeighbors, neighbor)
		}
	}
	return sshNeighbors
}
