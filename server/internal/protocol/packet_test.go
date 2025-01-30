package protocol

import (
	"reflect"
	"server/internal/protocol/proto_defs"
	"testing"
)

func TestPacket_MarshalUnmarshalBinary(t *testing.T) {

	packetHeader, _ := NewPacketHeader(
		PacketHeaderWithVersion(proto_defs.ProtocolV1),
		PacketHeaderWithMessageId(proto_defs.NewMessageId()),
		PacketHeaderWithMessageType(proto_defs.MessageTypeRequest),
		PacketHeaderWithPacketNumber(uint8(0)),
		PacketHeaderWithTotalPackets(uint8(1)),
		PacketHeaderWithFlags(
			proto_defs.FlagAckRequired,
		),
		PacketHeaderWithPayloadLength(uint16(10)),
	)

	packetPayload := make([]byte, 10)
	for i := 0; i < 10; i++ {
		packetPayload[i] = uint8(i + 1)
	}

	packet, _ := NewPacket(*packetHeader, packetPayload)

	packetBytes, _ := packet.MarshalBinary()

	regenPacket := &Packet{}
	_ = regenPacket.UnmarshalBinary(packetBytes)

	if !reflect.DeepEqual(regenPacket, packet) {
		t.Error("Packets do not match after marshalling/unmarshalling")
	}

}
