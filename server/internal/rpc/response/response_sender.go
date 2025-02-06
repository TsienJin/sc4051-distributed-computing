package response

import (
	"log/slog"
	"net"
	"server/internal/network"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
)

func SendResponse(c *net.UDPConn, a *net.UDPAddr, r *Response) {

	// Set response in history
	GetResponseHistoryInstance().AddResponse(r)

	// Create response message
	message, err := protocol.NewMessage(
		&protocol.PacketHeaderDistilled{
			Version:     proto_defs.ProtocolV1,
			MessageId:   proto_defs.NewMessageId(),
			MessageType: proto_defs.MessageTypeResponse,
			RequireAck:  true,
		},
		r,
	)
	if err != nil {
		slog.Error("Unable to create response message", "err", err)
		return
	}

	packets, err := message.ToPackets()
	if err != nil {
		slog.Error("Unable to create response message packets", "err", err)
		return
	}

	for _, p := range packets {
		if err := network.SendPacket(c, a, p); err != nil {
			slog.Error("Unable to send response message packet", "err", err)
		}
	}

}
