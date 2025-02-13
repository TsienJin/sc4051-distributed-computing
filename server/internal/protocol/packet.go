package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"server/internal/protocol/proto_defs"
)

// Packet s are the data structs used to represent the underlying data.
// there is a limit size of proto_def.PacketSizeLimit
type Packet struct {
	Header  PacketHeader
	Payload []byte

	Checksum uint32
}

func NewPacket(h PacketHeader, p []byte) (*Packet, error) {

	buf := &bytes.Buffer{}

	if err := binary.Write(buf, binary.BigEndian, h); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p); err != nil {
		return nil, err
	}

	return &Packet{
		Header:   h,
		Payload:  p,
		Checksum: MakeChecksum(buf.Bytes()),
	}, nil
}

func (p *Packet) ToBytes() ([]byte, error) {

	buf := &bytes.Buffer{}

	// Serialize the header
	if err := binary.Write(buf, binary.BigEndian, p.Header); err != nil {
		return nil, err
	}

	// Serialize the payload (write the bytes directly)
	if err := binary.Write(buf, binary.BigEndian, p.Payload); err != nil {
		return nil, err
	}

	// Serialize the checksum
	if err := binary.Write(buf, binary.BigEndian, p.Checksum); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Packet) MarshalBinary() ([]byte, error) {
	return p.ToBytes()
}

func (p *Packet) UnmarshalBinary(data []byte) error {

	// Handle header
	packetHeader := &PacketHeader{}
	if err := packetHeader.UnmarshalBinary(data[:proto_defs.PacketHeaderSize]); err != nil {
		return err
	}
	p.Header = *packetHeader

	// Handle payload data
	payloadData := make([]byte, p.Header.PayloadLength)
	copy(payloadData, data[proto_defs.PacketHeaderSize:proto_defs.PacketHeaderSize+p.Header.PayloadLength])
	p.Payload = payloadData

	// Handle checksum
	p.Checksum = GetChecksumFromChecksumBytes(data[proto_defs.PacketHeaderSize+p.Header.PayloadLength : proto_defs.PacketHeaderSize+p.Header.PayloadLength+proto_defs.PacketChecksumSize])

	// Validate checksum
	if !ValidateChecksum(data[:proto_defs.PacketHeaderSize+p.Header.PayloadLength], p.Checksum) {
		return errors.New("checksum does not match payload")
	}

	return nil
}
