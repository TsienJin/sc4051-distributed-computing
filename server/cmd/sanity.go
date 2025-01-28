package main

import (
	"fmt"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
)

func main() {

	headerDistilled := &protocol.PacketHeaderDistilled{
		Version:     proto_defs.ProtocolV1,
		MessageId:   proto_defs.NewMessageId(),
		MessageType: proto_defs.MessageTypeError,
		RequireAck:  true,
	}

	msg := protocol.NewMessage(headerDistilled, []byte{1, 2, 3})

	packets, _ := msg.ToPackets()

	for _, p := range packets {
		b, _ := p.ToBytes()
		fmt.Printf("% X\n", b)
	}

}
