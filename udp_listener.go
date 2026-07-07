package main

import (
	"fmt"
	"net"
	"time"
)

const (
	BufferSize           = 4096
	HealthCheckInterval  = 30 * time.Second
	HealthCheckTimeout   = 5 * time.Second
)

// UDPListener handles bidirectional UDP packet relay
type UDPListener struct {
	localAddr  *net.UDPAddr
	targetAddr *net.UDPAddr
	conn       *net.UDPConn
	clientMap  map[string]*net.UDPAddr // Track client addresses
	targetConn *net.UDPConn            // Connection to target server
	stopChan   chan bool
	IsHealthy  bool
}

// NewUDPListener creates a new UDP listener
func NewUDPListener(localAddrStr, targetAddrStr string) (*UDPListener, error) {
	localAddr, err := net.ResolveUDPAddr("udp", localAddrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local address: %w", err)
	}

	targetAddr, err := net.ResolveUDPAddr("udp", targetAddrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target address: %w", err)
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP: %w", err)
	}

	// Create a separate connection to the target server
	targetConn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to dial target server: %w", err)
	}

	return &UDPListener{
		localAddr:  localAddr,
		targetAddr: targetAddr,
		conn:       conn,
		targetConn: targetConn,
		clientMap:  make(map[string]*net.UDPAddr),
		stopChan:   make(chan bool),
		IsHealthy:  true,
	}, nil
}

// RelayPackets starts the main packet relay loop
func (ul *UDPListener) RelayPackets() error {
	buffer := make([]byte, BufferSize)

	for {
		select {
		case <-ul.stopChan:
			return nil
		default:
			// Set read timeout to allow periodic health checks
			ul.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

			n, remoteAddr, err := ul.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout is expected, loop continues
				}
				return fmt.Errorf("read error: %w", err)
			}

			// Register client
			ul.clientMap[remoteAddr.String()] = remoteAddr

			// Parse and log packet
			packet := ParseGamePacket(buffer[:n])
			fmt.Printf("[CLIENT %s] Received packet (%d bytes) | Type: %s\n", 
				remoteAddr.String(), n, packet.Type)

			// Forward to target server
			_, err = ul.targetConn.Write(buffer[:n])
			if err != nil {
				fmt.Printf("Error forwarding to target: %v\n", err)
				ul.IsHealthy = false
				continue
			}

			// Read response from target server
			ul.targetConn.SetReadDeadline(time.Now().Add(2 * time.Second))
			respBuffer := make([]byte, BufferSize)
			respN, err := ul.targetConn.Read(respBuffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					fmt.Printf("No response from target server (timeout)\n")
					ul.IsHealthy = false
					continue
				}
				fmt.Printf("Error reading from target: %v\n", err)
				ul.IsHealthy = false
				continue
			}

			ul.IsHealthy = true

			// Relay response back to client
			_, err = ul.conn.WriteToUDP(respBuffer[:respN], remoteAddr)
			if err != nil {
				fmt.Printf("Error sending response to client: %v\n", err)
				continue
			}

			fmt.Printf("[TARGET %s] Relayed response (%d bytes) back to client\n", 
				ul.targetAddr.String(), respN)
		}
	}
}

// StartHealthCheck periodically checks if the target server is alive
func (ul *UDPListener) StartHealthCheck() {
	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		ul.checkHealth()
	}
}

// checkHealth sends a ping to the target server
func (ul *UDPListener) checkHealth() {
	// Send a simple heartbeat packet (can be extended for specific game protocols)
	heartbeatPacket := []byte{0xFF, 0xFF, 0xFF, 0xFF} // Quake3 style heartbeat

	ul.targetConn.SetWriteDeadline(time.Now().Add(HealthCheckTimeout))
	_, err := ul.targetConn.Write(heartbeatPacket)
	if err != nil {
		fmt.Printf("[HEALTH CHECK] Failed to send heartbeat: %v\n", err)
		ul.IsHealthy = false
		return
	}

	ul.targetConn.SetReadDeadline(time.Now().Add(HealthCheckTimeout))
	respBuffer := make([]byte, BufferSize)
	_, err = ul.targetConn.Read(respBuffer)
	if err != nil {
		fmt.Printf("[HEALTH CHECK] Target server unhealthy: %v\n", err)
		ul.IsHealthy = false
		return
	}

	fmt.Printf("[HEALTH CHECK] Target server is healthy ✓\n")
	ul.IsHealthy = true
}

// Close closes the UDP listener
func (ul *UDPListener) Close() error {
	close(ul.stopChan)
	if ul.conn != nil {
		ul.conn.Close()
	}
	if ul.targetConn != nil {
		ul.targetConn.Close()
	}
	return nil
}

// GetClientCount returns the number of connected clients
func (ul *UDPListener) GetClientCount() int {
	return len(ul.clientMap)
}
