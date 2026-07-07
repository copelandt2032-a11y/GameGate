# GameGate — Smart UDP Game Proxy

A high-performance UDP game proxy built in Go for intercepting, logging, and relaying multiplayer game traffic. Perfect for DDoS protection, packet inspection, multi-region routing, and game server health monitoring.

## Features

✅ **UDP Packet Forwarding** — Listen on a local UDP port, forward traffic to the real game server, relay responses back to clients  
✅ **Heartbeat / Ping Monitor** — Periodic health checks to ensure the backend game server is alive  
✅ **Game Protocol Parsing** — Supports Source Engine (A2S queries), Minecraft server pings, and Quake-style packets  
✅ **Client Tracking** — Maintains a map of connected clients for analytics and filtering  
✅ **Graceful Shutdown** — Clean SIGINT/SIGTERM handling  

## Supported Game Protocols

- **Valve Source Engine** (CS:GO, TF2, etc.) — A2S_INFO, A2S_PLAYER, A2S_RULES, A2S_PING
- **Minecraft Java Edition** — Legacy and modern ping formats
- **Quake-style Games** — Generic 0xFFFFFFFF heartbeat packets

## Building & Running

### Prerequisites
- Go 1.21 or later

### Build
```bash
go build -o GameGate .
```

### Run
```bash
./GameGate
```

By default, the proxy listens on `0.0.0.0:27015` and forwards to `127.0.0.1:27016`.

#### Custom Configuration
Edit the `main()` function in `main.go` to change `LocalAddr` and `TargetAddr`:

```go
config := ProxyConfig{
    LocalAddr:  ":27015",           // Listen on this port
    TargetAddr: "192.168.1.100:27015", // Forward to this game server
}
```

## Architecture

```
Client 1 ──┐
Client 2 ──┤  [GameProxy UDP Listener] ──→ [Target Game Server]
Client 3 ──┘    • Packet parsing       ←── [Response]
           • Health check monitor
           • Client registry
```

### Core Components

1. **main.go** — Entry point, proxy initialization, signal handling
2. **udp_listener.go** — UDP socket management, bidirectional relay, health check logic
3. **game_protocol.go** — Protocol detection and packet parsing (A2S, Minecraft, Quake)

## Example Output

```
Starting gaming proxy on :27015 -> forwarding to 127.0.0.1:27016
Proxy is now listening for incoming packets...
[CLIENT 127.0.0.1:54321] Received packet (25 bytes) | Type: A2S_INFO
[TARGET 127.0.0.1:27016] Relayed response (256 bytes) back to client
[HEALTH CHECK] Target server is healthy ✓
```

## Configuration & Customization

### Health Check Interval
Modify `HealthCheckInterval` in `udp_listener.go` (default: 30 seconds):

```go
const HealthCheckInterval = 30 * time.Second
```

### Buffer Size
Increase `BufferSize` for games with large packets (default: 4096 bytes):

```go
const BufferSize = 4096
```

### Custom Protocol Parsing
Add new game protocols in `game_protocol.go`:

1. Add a detection function (`isXXXGame()`)
2. Implement a parse function (`parseXXXPacket()`)
3. Register in `ParseGamePacket()`

## Testing

### Test with nc (netcat)
```bash
# Terminal 1: Start the proxy
./GameGate

# Terminal 2: Send a test packet
echo "test" | nc -u localhost 27015
```

### Test with a Local Game Server
Run a local game server on `127.0.0.1:27016`, then point clients to the proxy at `127.0.0.1:27015`.

## Future Enhancements

- 🔄 Load balancing across multiple backend servers
- 📊 Packet statistics and analytics dashboard
- 🔒 Packet filtering and rate limiting
- 🌍 Multi-region failover support
- 📝 Structured logging (JSON, syslog)
- 🐳 Docker containerization
- 🔐 Optional packet encryption/obfuscation

## License

MIT License — Feel free to use and modify!

## Contributing

Contributions are welcome! Fork the repo and submit a PR.

---

**Built with ❤️ for game developers and network engineers**
