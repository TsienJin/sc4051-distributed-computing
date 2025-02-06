package handle

import (
	"log/slog"
	"net"
	"server/internal/handle/handle_requests"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
)

func IncomingMessage(c *net.UDPConn, a *net.UDPAddr, m *protocol.Message) {

	slog.Info("Handling message", "MessageType", m.Header.MessageType, "MessageId", m.Header.MessageId)

	switch m.Header.MessageType {
	case proto_defs.MessageTypeRequest:
		handle_requests.Sort(c, a, m)
		break
	default:
		slog.Error("Message type not supported yet", "MessageType", m.Header.MessageType)
	}
}
