package network

import (
	"errors"
	"server/internal/protocol"
	"sync"
)

// History is responsible for keeping track of all previously sent messages that require acknowledgement.
type History struct {
	messages map[protocol.PacketIdent]*protocol.Packet
}

var instance *History
var once sync.Once

func GetHistoryInstance() *History {
	once.Do(func() {
		instance = &History{
			messages: make(map[protocol.PacketIdent]*protocol.Packet),
		}
	})
	return instance
}

func (h *History) Append(p *protocol.Packet) {
	h.messages[protocol.ExtractIdentFromPacket(p)] = p
}

func (h *History) Remove(i protocol.PacketIdent) {
	delete(h.messages, i)
}

func (h *History) Get(i protocol.PacketIdent) (*protocol.Packet, error) {
	if p, exists := h.messages[i]; !exists {
		return nil, errors.New("corresponding packet does not exist")
	} else {
		return p, nil
	}
}
