package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"server/internal/protocol/proto_defs"
)

type PacketHeaderDistilled struct {
	Version     proto_defs.ProtocolVersion
	MessageId   proto_defs.MessageId
	MessageType proto_defs.MessageType
	RequireAck  bool
}

type PacketHeader struct {
	Version       proto_defs.ProtocolVersion
	MessageId     proto_defs.MessageId
	MessageType   proto_defs.MessageType
	PacketNumber  uint8
	TotalPackets  uint8
	Flags         proto_defs.Flags
	PayloadLength uint16
}

type PacketHeaderOption func(header *PacketHeader)

func PacketHeaderFromDistilled(v *PacketHeaderDistilled) PacketHeaderOption {
	return func(header *PacketHeader) {
		header.Version = v.Version
		header.MessageId = v.MessageId
		header.MessageType = v.MessageType
		if v.RequireAck {
			header.Flags = proto_defs.NewFlags(proto_defs.FlagAckRequired)
		}
	}
}

func PacketHeaderWithVersion(v proto_defs.ProtocolVersion) PacketHeaderOption {
	return func(header *PacketHeader) {
		header.Version = v
	}
}

func PacketHeaderWithMessageId(v proto_defs.MessageId) PacketHeaderOption {
	return func(header *PacketHeader) {
		header.MessageId = v
	}
}

func PacketHeaderWithMessageType(v proto_defs.MessageType) PacketHeaderOption {
	return func(header *PacketHeader) {
		header.MessageType = v
	}
}

func PacketHeaderWithPacketNumber(v uint8) PacketHeaderOption {
	return func(header *PacketHeader) {
		header.PacketNumber = v
	}
}

func PacketHeaderWithTotalPackets(v uint8) PacketHeaderOption {
	return func(header *PacketHeader) {
		header.TotalPackets = v
	}
}

func PacketHeaderWithFlags(v proto_defs.Flags) PacketHeaderOption {
	return func(header *PacketHeader) {
		header.Flags = proto_defs.NewFlags(header.Flags, v)
	}
}

func PacketHeaderWithPayloadLength(v uint16) PacketHeaderOption {
	return func(header *PacketHeader) {
		header.PayloadLength = v
	}
}

func NewPacketHeader(opts ...PacketHeaderOption) (*PacketHeader, error) {
	p := &PacketHeader{}
	for _, o := range opts {
		o(p)
	}

	if err := p.validate(); err != nil {
		return nil, err
	}

	return p, nil
}

// validate checks for required fields.
// - Version
// - MessageId
// - MessageType
// - TotalPackets
// Remaining fields are somewhat optional or can intrinsically be falsy
func (p *PacketHeader) validate() error {
	if p.Version == 0 {
		return errors.New("packet Header version not set")
	}

	if p.MessageId == [16]byte{} {
		return errors.New("packet Header message id not set")
	}

	if p.MessageType == 0 {
		return errors.New("packet Header message type not set")
	}

	if p.TotalPackets == 0 {
		return errors.New("packet Header total packets not set")
	}

	return nil
}

func (p *PacketHeader) MarshalBinary() ([]byte, error) {

	buf := make([]byte, proto_defs.PacketHeaderSize)

	buf[0] = uint8(p.Version)
	copy(buf[1:], p.MessageId[:])
	buf[17] = uint8(p.MessageType)
	buf[18] = p.PacketNumber
	buf[19] = p.TotalPackets
	buf[20] = uint8(p.Flags)
	binary.BigEndian.PutUint16(buf[21:], p.PayloadLength)

	return buf, nil
}

func (p *PacketHeader) UnmarshalBinary(data []byte) error {
	// Ensure the input data is at least the expected size
	if len(data) < proto_defs.PacketHeaderSize {
		return fmt.Errorf("insufficient data to unmarshal PacketHeader: expected %d bytes, got %d", proto_defs.PacketHeaderSize, len(data))
	}

	// Extract fields from the byte array
	p.Version = proto_defs.ProtocolVersion(data[0])      // First byte: Protocol version
	copy(p.MessageId[:], data[1:17])                     // Next 16 bytes: Message ID
	p.MessageType = proto_defs.MessageType(data[17])     // Message type
	p.PacketNumber = data[18]                            // Packet number
	p.TotalPackets = data[19]                            // Total packets
	p.Flags = proto_defs.Flags(data[20])                 // Flags
	p.PayloadLength = binary.BigEndian.Uint16(data[21:]) // PayloadLength (last 2 bytes)

	return nil
}

func (p *PacketHeader) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
