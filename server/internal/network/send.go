package network

import (
	"log/slog"
	"net"
	"server/internal/chance"
	"server/internal/protocol"
)

// SendPacket is responsible for sending the packet to the given address.
func SendPacket(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) error {

	// Chance event to drop sent packet
	if chance.DropPacket() {
		slog.Info("[OUT:DROPPED] Dropping packet, simulated network error", "target", a.String())
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
	return nil
}
