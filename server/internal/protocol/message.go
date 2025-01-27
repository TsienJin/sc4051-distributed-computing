package protocol

import "server/internal/protocol/proto_defs"

type Message struct {
	header  *PacketHeaderDistilled
	payload []byte
}

func NewMessage(
	header *PacketHeaderDistilled,
	payload []byte,
) *Message {
	return &Message{
		header:  header,
		payload: payload,
	}
}

func (m *Message) ToPackets() ([]*Packet, error) {

	totalPayloadSize := len(m.payload)

	// Determine number of packets needed to send
	nPackets := totalPayloadSize / proto_defs.PacketPayloadSizeLimit
	nBytesRemainder := totalPayloadSize % proto_defs.PacketPayloadSizeLimit
	if nBytesRemainder != 0 {
		nPackets++
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

		// Create packet header
		packetHeader, err := NewPacketHeader(
			PacketHeaderFromDistilled(m.header),
			PacketHeaderWithPacketNumber(uint16(i)),
			PacketHeaderWithTotalPackets(uint8(nPackets)),
			PacketHeaderWithPayloadLength(uint16(len(data))),
		)
		if err != nil {
			return nil, err
		}

		// Add packet to array
		packets[i] = NewPacket(*packetHeader, data)
	}

	return packets, nil

}
