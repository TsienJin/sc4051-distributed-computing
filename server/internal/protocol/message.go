package protocol

import (
	"encoding"
	"log/slog"
	"server/internal/protocol/proto_defs"
)

type Message struct {
	header  *PacketHeaderDistilled
	payload []byte
}

func NewMessageFromBytes(
	header *PacketHeaderDistilled,
	payload []byte,
) *Message {
	return &Message{
		header:  header,
		payload: payload,
	}
}

func NewMessage(
	header *PacketHeaderDistilled,
	p encoding.BinaryMarshaler,
) (*Message, error) {
	data, err := p.MarshalBinary()
	if err != nil {
		slog.Error("Unable to marshal payload into bytes", "Payload", p)
		return nil, err
	}
	return NewMessageFromBytes(header, data), nil
}

func (m *Message) ToPackets() ([]*Packet, error) {

	totalPayloadSize := len(m.payload)

	// Determine number of packets needed to send
	nPackets := totalPayloadSize / proto_defs.PacketPayloadSizeLimit
	nBytesRemainder := totalPayloadSize % proto_defs.PacketPayloadSizeLimit
	if nBytesRemainder != 0 {
		nPackets++
	}

	// Manage flags for packets
	packetFlags := proto_defs.NewFlags()
	if nPackets > 1 {
		packetFlags = proto_defs.NewFlags(packetFlags, proto_defs.FlagFragment)
	}

	packets := make([]*Packet, nPackets)

	// Generate all the packets for the response
	for i := 0; i < nPackets; i++ {

		var data []byte

		// Get aligned packet data
		if i == nPackets-1 {
			data = make([]byte, nBytesRemainder)
			copy(data, m.payload[i*proto_defs.PacketPayloadSizeLimit:])
		} else {
			data = make([]byte, proto_defs.PacketPayloadSizeLimit)
			leftLimit := i * proto_defs.PacketPayloadSizeLimit
			rightLimit := leftLimit + proto_defs.PacketPayloadSizeLimit
			copy(data, m.payload[leftLimit:rightLimit])
		}

		// Create packet Header
		packetHeader, err := NewPacketHeader(
			PacketHeaderFromDistilled(m.header),
			PacketHeaderWithFlags(packetFlags),
			PacketHeaderWithPacketNumber(uint8(i)),
			PacketHeaderWithTotalPackets(uint8(nPackets)),
			PacketHeaderWithPayloadLength(uint16(len(data))),
		)
		if err != nil {
			return nil, err
		}

		// Add packet to array
		p, err := NewPacket(*packetHeader, data)
		if err != nil {
			return nil, err
		}
		packets[i] = p
	}

	return packets, nil

}
