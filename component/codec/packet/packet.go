package packet

import (
	"fmt"
	"github.com/pkg/errors"
)

// Type represents the network packet's type such as: handshake and so on.
type Type byte

const (
	_ Type = iota
	// Handshake represents a handshake: request(client) <====> handshake response(server)
	Handshake = 0x01

	// HandshakeAck represents a handshake ack from client to server
	HandshakeAck = 0x02

	// Heartbeat represents a heartbeat
	Heartbeat = 0x03

	// Data represents a common data packet
	Data = 0x04

	// Kick represents a kick off packet
	Kick = 0x05 // disconnect message from server
)

// ErrWrongPacketType represents a wrong packet type.
var ErrWrongPacketType = errors.New("wrong packet type")

// Packet represents a network packet.
type Packet struct {
	Type   Type
	Length int
	Data   []byte
}

//String represents the Packet's in text mode.
func (p *Packet) String() string {
	return fmt.Sprintf("Type: %d, Length: %d, Data: %v", p.Type, p.Length, p.Data)
}
