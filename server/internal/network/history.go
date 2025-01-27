package network

import (
	"errors"
	"server/internal/protocol"
	"sync"
)

// History is responsible for keeping track of all previously sent messages that require acknowledgement.
type History struct {
	sync.RWMutex
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
	h.Lock()
	defer h.Unlock()
	h.messages[protocol.ExtractIdentFromPacket(p)] = p
}

func (h *History) Remove(i protocol.PacketIdent) {
	h.Lock()
	defer h.Unlock()
	delete(h.messages, i)
}

func (h *History) Get(i protocol.PacketIdent) (*protocol.Packet, error) {
	h.RLock()
	defer h.RUnlock()
	if p, exists := h.messages[i]; !exists {
		return nil, errors.New("corresponding packet does not exist")
	} else {
		return p, nil
	}
}
