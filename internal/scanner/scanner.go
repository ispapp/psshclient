package scanner

import (
	"context"
	"fmt"
	"ispappclient/pkg/gomap"
	"net"
	"sync"
	"time"
)

// Device represents a discovered device
type Device struct {
	IP        string
	Hostname  string
	Port22    bool
	Port23    bool
	SSHPort   int // SSH port
	Status    string
	Username  string // SSH username
	Password  string // SSH password
	Connected bool   // SSH connection status
}

// PortResult represents the structure that gomap returns for each port
// This is based on the gomap library's internal structure
type PortResult struct {
	Port    int
	State   string
	Service string
}

// ScanSubnet scans a subnet for devices with SSH (22) or Telnet (23) ports open
func ScanSubnet(ctx context.Context, subnet string, progressCallback func(string)) ([]Device, error) {
	var devices []Device
	var devicesMutex sync.Mutex

	// Parse subnet
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet format: %v", err)
	}

	// Generate IP list from subnet
	ips := generateIPList(ipNet)

	progressCallback(fmt.Sprintf("Scanning %d hosts in %s...", len(ips), subnet))

	// Channel to control concurrent scans
	const maxConcurrent = 50 // Limit concurrent scans to avoid overwhelming the network
	semaphore := make(chan struct{}, maxConcurrent)

	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Progress tracking
	var scannedCount int
	var progressMutex sync.Mutex

	// Scan each IP concurrently
	for _, ip := range ips {
		// Check for cancellation before starting new goroutine
		select {
		case <-ctx.Done():
			progressCallback("Scan cancelled")
			wg.Wait() // Wait for existing goroutines to finish
			return devices, ctx.Err()
		default:
		}

		wg.Add(1)

		go func(currentIP string) {
			defer wg.Done()

			// Acquire semaphore (limit concurrency)
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Check for cancellation again inside goroutine
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Update progress
			progressMutex.Lock()
			scannedCount++
			currentCount := scannedCount
			progressMutex.Unlock()

			progressCallback(fmt.Sprintf("Scanning %s (%d/%d)", currentIP, currentCount, len(ips)))

			// Use gomap to scan this specific IP with timeout
			result, err := gomap.ScanIP(currentIP, "tcp", true, false) // fastscan=true, stealth=false

			if err != nil {
				// Check if the main context was cancelled
				if ctx.Err() != nil {
					return
				}
				return // Skip hosts that can't be scanned or timed out
			}

			// Check if any ports are open
			if len(result.Results) == 0 {
				return // No open ports, skip this host
			}

			device := Device{
				IP:       currentIP,
				Hostname: result.Hostname,
				Status:   "Up",
			}

			// Check which ports are open (SSH=22, Telnet=23)
			hasSSHOrTelnet := false
			for _, portResult := range result.Results {
				if portResult.State {
					if portResult.Port == 22 {
						device.Port22 = true
						device.SSHPort = 22 // Default SSH port
						hasSSHOrTelnet = true
					}
					if portResult.Port == 23 {
						device.Port23 = true
						hasSSHOrTelnet = true
					}
				}
				if portResult.State {
					if portResult.Port == 22 {
						device.Port22 = true
						hasSSHOrTelnet = true
					}
					if portResult.Port == 23 {
						device.Port23 = true
						hasSSHOrTelnet = true
					}
				}
			}

			// Only add devices that have SSH or Telnet open
			if hasSSHOrTelnet {
				devicesMutex.Lock()
				devices = append(devices, device)
				devicesMutex.Unlock()
			}
		}(ip)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check if scan was cancelled
	if ctx.Err() != nil {
		progressCallback("Scan cancelled")
		return devices, ctx.Err()
	}

	progressCallback(fmt.Sprintf("Scan complete. Found %d devices with SSH/Telnet.", len(devices)))
	return devices, nil
}

// generateIPList generates a list of IP addresses from a CIDR subnet
func generateIPList(ipNet *net.IPNet) []string {
	var ips []string

	// Get network address and mask
	ip := ipNet.IP.To4()
	if ip == nil {
		return ips // IPv6 not supported in this simple implementation
	}

	mask := ipNet.Mask
	network := ip.Mask(mask)

	// Calculate broadcast address
	broadcast := make(net.IP, len(network))
	copy(broadcast, network)
	for i := range network {
		broadcast[i] |= ^mask[i]
	}

	// Generate all IPs in range (excluding network and broadcast)
	start := ipToInt(network)
	end := ipToInt(broadcast)

	for i := start + 1; i < end; i++ {
		ips = append(ips, intToIP(i).String())
	}

	return ips
}

// ipToInt converts an IP address to integer
func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
}

// intToIP converts an integer to IP address
func intToIP(i uint32) net.IP {
	return net.IPv4(byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

// ScanSingleHost scans a single host for SSH/Telnet ports
func ScanSingleHost(ctx context.Context, ip string) (*Device, error) {
	device := &Device{
		IP:     ip,
		Status: "Down",
	}

	// Create a timeout context for the scan
	scanCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Use gomap to scan this specific IP with timeout
	result, err := scanIPWithTimeout(scanCtx, ip)
	if err != nil {
		return device, err
	}

	// Set hostname from gomap result
	device.Hostname = result.Hostname
	if device.Hostname == "" {
		device.Hostname = ip
	}

	// Check if any ports are open
	if len(result.Results) > 0 {
		device.Status = "Up"

		// Check which ports are open (SSH=22, Telnet=23)
		for _, portResult := range result.Results {
			if portResult.State {
				if portResult.Port == 22 {
					device.Port22 = true
				}
				if portResult.Port == 23 {
					device.Port23 = true
				}
			}
		}
	}

	return device, nil
}

// scanIPWithTimeout scans an IP address with a timeout context
func scanIPWithTimeout(ctx context.Context, ip string) (*gomap.IPScanResult, error) {
	// Channel to receive the scan result
	resultChan := make(chan *gomap.IPScanResult, 1)
	errorChan := make(chan error, 1)

	// Start the scan in a goroutine
	go func() {
		result, err := gomap.ScanIP(ip, "tcp", true, false) // fastscan=true, stealth=false
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for either the scan to complete or the context to be cancelled/timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}
