package interfaces

import "server/internal/protocol"

type MessageInterface interface {
	ToPackets() ([]*protocol.Packet, error)
}
