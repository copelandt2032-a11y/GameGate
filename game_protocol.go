package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// GamePacket represents a parsed game protocol packet
type GamePacket struct {
	Type     string
	Protocol string
	Data     []byte
	Raw      []byte
}

// ParseGamePacket attempts to identify and parse a game packet
func ParseGamePacket(data []byte) *GamePacket {
	packet := &GamePacket{
		Type:     "UNKNOWN",
		Protocol: "UNKNOWN",
		Data:     data,
		Raw:      data,
	}

	if len(data) < 4 {
		return packet
	}

	// Check for Source Engine A2S_INFO query (Valve games)
	if isA2SQuery(data) {
		packet.Type = "A2S_QUERY"
		packet.Protocol = "Source Engine"
		parseA2SInfo(packet, data)
		return packet
	}

	// Check for Minecraft server ping
	if isMinecraftPing(data) {
		packet.Type = "MINECRAFT_PING"
		packet.Protocol = "Minecraft"
		return packet
	}

	// Check for Quake-style heartbeat
	if isQuakeHeartbeat(data) {
		packet.Type = "HEARTBEAT"
		packet.Protocol = "Quake"
		return packet
	}

	// Default: generic packet
	packet.Type = "DATA"
	packet.Protocol = "Generic UDP"

	return packet
}

// isA2SQuery checks if this is a Valve A2S_INFO query (0xFFFFFFFF + command)
func isA2SQuery(data []byte) bool {
	if len(data) < 5 {
		return false
	}

	// A2S packets start with 0xFFFFFFFF
	if data[0] == 0xFF && data[1] == 0xFF && data[2] == 0xFF && data[3] == 0xFF {
		cmd := data[4]
		// Common A2S commands: 'T' (0x54) for info, 'U' (0x55) for player list, 'R' (0x52) for rules
		return cmd == 0x54 || cmd == 0x55 || cmd == 0x52 || cmd == 0x57 // 'W' for ping
	}

	return false
}

// parseA2SInfo extracts information from A2S_INFO response
func parseA2SInfo(packet *GamePacket, data []byte) {
	if len(data) < 5 {
		return
	}

	command := data[4]
	switch command {
	case 0x54: // A2S_INFO
		packet.Type = "A2S_INFO"
		packet.Data = data[5:]
	case 0x55: // A2S_PLAYER
		packet.Type = "A2S_PLAYER"
		packet.Data = data[5:]
	case 0x52: // A2S_RULES
		packet.Type = "A2S_RULES"
		packet.Data = data[5:]
	case 0x57: // A2S_PING
		packet.Type = "A2S_PING"
		packet.Data = data[5:]
	}
}

// isMinecraftPing checks for Minecraft server list ping (protocol 0xFE for legacy or modern handshake)
func isMinecraftPing(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// Legacy ping: 0xFE + 0x01
	if data[0] == 0xFE && data[1] == 0x01 {
		return true
	}

	// Modern handshake: 0x00 (packet length) followed by 0x00 (handshake)
	if len(data) >= 2 && data[0] == 0x00 && data[1] == 0x00 {
		return true
	}

	return false
}

// isQuakeHeartbeat checks for Quake-style packets (0xFFFFFFFF)
func isQuakeHeartbeat(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	return data[0] == 0xFF && data[1] == 0xFF && data[2] == 0xFF && data[3] == 0xFF
}

// ParseA2SInfoResponse parses the A2S_INFO response to extract server details
func ParseA2SInfoResponse(data []byte) map[string]string {
	info := make(map[string]string)

	if len(data) < 6 {
		return info
	}

	// Response header: 0xFFFFFFFF (4 bytes) + 0x49 (1 byte for 'I')
	if !(data[0] == 0xFF && data[1] == 0xFF && data[2] == 0xFF && data[3] == 0xFF && data[4] == 0x49) {
		return info
	}

	buf := bytes.NewBuffer(data[5:])

	// Protocol version (1 byte)
	var protocol byte
	binary.Read(buf, binary.LittleEndian, &protocol)
	info["protocol"] = fmt.Sprintf("%d", protocol)

	// Server name (null-terminated string)
	serverName, _ := readCString(buf)
	info["server_name"] = serverName

	// Map name (null-terminated string)
	mapName, _ := readCString(buf)
	info["map"] = mapName

	// Game folder (null-terminated string)
	gameFolder, _ := readCString(buf)
	info["game"] = gameFolder

	// Game description (null-terminated string)
	gameDesc, _ := readCString(buf)
	info["description"] = gameDesc

	return info
}

// readCString reads a null-terminated string from a buffer
func readCString(buf *bytes.Buffer) (string, error) {
	var result []byte
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return string(result), err
		}
		if b == 0 {
			break
		}
		result = append(result, b)
	}
	return string(result), nil
}

// FormatPacketLog returns a formatted log line for packet info
func FormatPacketLog(packet *GamePacket, size int) string {
	var details []string

	details = append(details, fmt.Sprintf("Type: %s", packet.Type))
	details = append(details, fmt.Sprintf("Protocol: %s", packet.Protocol))
	details = append(details, fmt.Sprintf("Size: %d bytes", size))

	// Include hex dump for debugging
	if len(packet.Data) > 0 {
		hexStr := fmt.Sprintf("%02X", packet.Data[:min(len(packet.Data), 16)])
		details = append(details, fmt.Sprintf("Data: %s...", hexStr))
	}

	return strings.Join(details, " | ")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
