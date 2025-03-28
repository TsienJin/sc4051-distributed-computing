package network

import (
	"log/slog"
	"net"
	"server/internal/chance"
	"server/internal/monitor"
	"server/internal/protocol"
)

// SendPacket is responsible for sending the packet to the given address.
func SendPacket(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) error {

	monitor.MarkPacketOut()

	// Chance event to drop sent packet
	if chance.DropPacket() {
		monitor.MarkPacketOutDropped()
		slog.Info("[OUT:DROPPED] Dropping packet, simulated network error", "target", a.String(), "packet_type", p.Header.MessageType)
		return nil
	}

	data, err := p.ToBytes()
	if err != nil {
		return err
	}
	if _, errSend := c.WriteToUDP(data, a); errSend != nil {
		return errSend
	}

	GetSendHistoryInstance().Append(c, a, p)

	slog.Info("[OUT:SUCCESS] Packet was successfully sent", "target", a.String(), "packet_type", p.Header.MessageType)
	return nil
}
