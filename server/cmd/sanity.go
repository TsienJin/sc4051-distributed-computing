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

	msg := protocol.NewMessageFromBytes(headerDistilled, []byte{1, 2, 3})

	packets, _ := msg.ToPackets()

	var headerUn protocol.PacketHeader

	for _, p := range packets {
		b, _ := p.ToBytes()
		hb, _ := p.Header.MarshalBinary()

		fmt.Printf("Header\t\t% X\n", hb)
		fmt.Printf("Everything\t% X\n", b)
		_ = headerUn.UnmarshalBinary(hb)
		fmt.Printf("Header Un\t%v\n", headerUn)
	}

}
