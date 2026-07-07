package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Proxy Config - Copilot loves explicit configuration structs
type ProxyConfig struct {
	LocalAddr  string // e.g., "0.0.0.0:27015"
	TargetAddr string // e.g., "12.34.56.78:27015"
}

type GameProxy struct {
	Config ProxyConfig
}

// Start kicks off the UDP proxy loop
func (p *GameProxy) Start() error {
	fmt.Printf("Starting gaming proxy on %s -> forwarding to %s\n", p.Config.LocalAddr, p.Config.TargetAddr)
	
	// Initialize UDP listener
	listener, err := NewUDPListener(p.Config.LocalAddr, p.Config.TargetAddr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	defer listener.Close()

	// Start health check monitor
	go listener.StartHealthCheck()

	// Start packet relay loop
	fmt.Println("Proxy is now listening for incoming packets...")
	return listener.RelayPackets()
}

func main() {
	config := ProxyConfig{
		LocalAddr:  ":27015",
		TargetAddr: "127.0.0.1:27016", // Your local test game server
	}

	proxy := &GameProxy{Config: config}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		os.Exit(0)
	}()

	if err := proxy.Start(); err != nil {
		log.Fatalf("Proxy error: %v", err)
	}
}
