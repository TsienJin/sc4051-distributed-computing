package network

import (
	"net"
	"server/internal/protocol"
)

// SendPacket is responsible for sending the packet to the given address.
func SendPacket(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) error {

	data, err := p.ToBytes()
	if err != nil {
		return err
	}
	if _, errSend := c.WriteToUDP(data, a); errSend != nil {
		return errSend
	}
	GetSendHistoryInstance().Append(p)
	return nil
}
