package protocol

import (
	"server/internal/protocol/proto_defs"
	"testing"
)

func TestNewMessage(t *testing.T) {
	distilledHeader := &PacketHeaderDistilled{
		Version:     proto_defs.ProtocolV1,
		MessageId:   proto_defs.NewMessageId(),
		MessageType: proto_defs.MessageTypeRequest,
		RequireAck:  false,
	}

	data := make([]byte, proto_defs.PacketPayloadSizeLimit+1)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i + 1)
	}

	msg := NewMessage(distilledHeader, data)

	packets, _ := msg.ToPackets()

	// Tests for packet "overflow"
	if len(packets) != 2 {
		t.Errorf("expcted 2 packets, received: %d", len(packets))
	}

	// Checks that the first payload is fully utilised
	if len(packets[0].Payload) != proto_defs.PacketPayloadSizeLimit {
		t.Error("did not adhere to packet payload size limit")
	}
}
