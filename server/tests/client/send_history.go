package client

import (
	"errors"
	"server/internal/protocol"
	"sync"
)

type sendHistory struct {
	mu      sync.RWMutex
	history map[protocol.PacketIdent]*protocol.Packet
}

func newSendHistory() *sendHistory {
	return &sendHistory{
		mu:      sync.RWMutex{},
		history: make(map[protocol.PacketIdent]*protocol.Packet),
	}
}

func (s *sendHistory) set(p *protocol.Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history[protocol.ExtractIdentFromPacket(p)] = p
}

func (s *sendHistory) clear(i *protocol.PacketIdent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.history, *i)
}

func (s *sendHistory) get(i *protocol.PacketIdent) (*protocol.Packet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, exists := s.history[*i]; exists {
		return p, nil
	} else {
		return nil, errors.New("packet does not exist")
	}
}
