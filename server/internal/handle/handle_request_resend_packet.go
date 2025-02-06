package handle

import (
	"log/slog"
	"net"
	"server/internal/network"
	"server/internal/protocol"
)

func RequestResendPacket(c *net.UDPConn, a *net.UDPAddr, m *protocol.Packet) {

	var p protocol.AckResendPayload
	if err := p.UnmarshalBinary(m.Payload); err != nil {
		slog.Error("Unable to unmarshal resend message payload", "err", err)
		return
	}

	h := network.GetSendHistoryInstance()
	packet, err := h.Get(*p.ToPacketIdent())
	if err != nil {
		slog.Error("Unable to retrieve corresponding packet from packet history", "err", err)
		return
	}
	err = network.SendPacket(c, a, packet)
	if err != nil {
		slog.Error("Unable to send packet", "err", err)
		return
	}

	slog.Info("Requested packet has been resent")
}
