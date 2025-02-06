package network

import (
	"errors"
	"log/slog"
	"net"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"server/internal/vars"
	"sync"
	"time"
)

type SendHistoryRecord struct {
	sync.RWMutex
	Conn    *net.UDPConn
	Addr    *net.UDPAddr
	Packet  *protocol.Packet
	Updated time.Time
}

func NewSendHistoryRecord(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) *SendHistoryRecord {
	return &SendHistoryRecord{
		Conn:    c,
		Addr:    a,
		Packet:  p,
		Updated: time.Now(),
	}
}

func (s *SendHistoryRecord) ResendPacket() {
	slog.Info("Resending packet")
	packet := s.GetPacket()
	if err := SendPacket(s.Conn, s.Addr, packet); err != nil {
		slog.Error("Unable to resend historical packet", "err", err)
	}
}

func (s *SendHistoryRecord) GetPacket() *protocol.Packet {
	s.Lock()
	defer s.Unlock()
	s.Updated = time.Now()
	return s.Packet
}

func (s *SendHistoryRecord) GetTime() *time.Time {
	s.RLock()
	defer s.RUnlock()
	return &s.Updated
}

// SendHistory is responsible for keeping track of all previously sent messages that require acknowledgement.
type SendHistory struct {
	sync.RWMutex
	messages map[protocol.PacketIdent]*SendHistoryRecord
}

var instance *SendHistory
var once sync.Once

func GetSendHistoryInstance() *SendHistory {
	once.Do(func() {
		instance = &SendHistory{
			messages: make(map[protocol.PacketIdent]*SendHistoryRecord),
		}

		// Create ticker to request outdated packets
		t := time.NewTicker(time.Duration(vars.GetStaticEnv().PacketReceiveTimeout) * time.Millisecond)
		go func() {
			defer t.Stop()
			for range t.C {
				instance.ResendUnAckPackets()
			}
		}()
	})
	return instance
}

func (h *SendHistory) ResendUnAckPackets() {
	h.Lock()
	defer h.Unlock()

	if len(h.messages) == 0 {
		return
	}

	slog.Info("Resending unacknowledged packets")
	historyCutoff := time.Now().Add(time.Duration(vars.GetStaticEnv().PacketTTL) * time.Millisecond)
	cutOffTime := time.Now().Add(-time.Duration(vars.GetStaticEnv().PacketReceiveTimeout) * time.Millisecond)

	for ident, p := range h.messages {

		// Check is packet is expired
		if p.GetTime().Before(historyCutoff) {
			slog.Info("Deleting expired packet", "Ident", ident)
			delete(h.messages, ident)
			continue
		}

		// Check if update time has been more than timeout and requires an ack
		if p.GetTime().Before(cutOffTime) && p.Packet.Header.Flags.AckRequired() {
			go p.ResendPacket()
		}
	}

}

func (h *SendHistory) Append(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) {

	// Do not add ack packets to history
	if p.Header.MessageType == proto_defs.MessageTypeAcknowledge {
		return
	}

	h.Lock()
	defer h.Unlock()
	h.messages[protocol.ExtractIdentFromPacket(p)] = NewSendHistoryRecord(c, a, p)
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
		return p.GetPacket(), nil
	}
}
