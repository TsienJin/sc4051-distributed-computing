package network

import (
	"errors"
	"server/internal/protocol"
	"sync"
)

// SendHistory is responsible for keeping track of all previously sent messages that require acknowledgement.
type SendHistory struct {
	sync.RWMutex
	messages map[protocol.PacketIdent]*protocol.Packet
}

var instance *SendHistory
var once sync.Once

func GetSendHistoryInstance() *SendHistory {
	once.Do(func() {
		instance = &SendHistory{
			messages: make(map[protocol.PacketIdent]*protocol.Packet),
		}
	})
	return instance
}

func (h *SendHistory) Append(p *protocol.Packet) {
	h.Lock()
	defer h.Unlock()
	h.messages[protocol.ExtractIdentFromPacket(p)] = p
}

func (h *SendHistory) Remove(i protocol.PacketIdent) {
	h.Lock()
	defer h.Unlock()
	delete(h.messages, i)
}

func (h *SendHistory) Get(i protocol.PacketIdent) (*protocol.Packet, error) {
	h.RLock()
	defer h.RUnlock()
	if p, exists := h.messages[i]; !exists {
		return nil, errors.New("corresponding packet does not exist")
	} else {
		return p, nil
	}
}
