package protocol

import (
	"github.com/google/go-cmp/cmp"
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
	if err := regenPacket.UnmarshalBinary(packetBytes); err != nil {
		t.Error(err)
	}

	if !cmp.Equal(regenPacket, packet) {
		t.Error("Packets do not match after marshalling/unmarshalling")
	}

	packetBytes[len(packetBytes)-1]++
	var faultyPacket Packet
	if err := faultyPacket.UnmarshalBinary(packetBytes); err == nil {
		t.Error("Expected error due to modified checksum")
	}

}
