package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"sync"
	"time"
)

const (
	PACKET_TTL    = time.Duration(15000) * time.Millisecond
	PACKET_RESEND = time.Duration(50) * time.Millisecond
)

type packetHistoryRecord struct {
	conn    *net.UDPConn
	addr    *net.UDPAddr
	packet  *protocol.Packet
	created time.Time
	updated time.Time
}

func newPacketHistoryRecord(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) *packetHistoryRecord {
	now := time.Now()
	return &packetHistoryRecord{
		conn:    c,
		addr:    a,
		packet:  p,
		created: now,
		updated: now,
	}
}

type sendManager struct {
	wg             sync.WaitGroup
	mu             sync.RWMutex
	ctx            context.Context
	cancelFunc     context.CancelFunc
	history        map[protocol.PacketIdent]*packetHistoryRecord
	requestHistory map[protocol.PacketIdent]*packetHistoryRecord
}

func newSendHistory() *sendManager {

	ctx, cancel := context.WithCancel(context.Background())

	s := &sendManager{
		wg:             sync.WaitGroup{},
		mu:             sync.RWMutex{},
		ctx:            ctx,
		cancelFunc:     cancel,
		history:        make(map[protocol.PacketIdent]*packetHistoryRecord),
		requestHistory: make(map[protocol.PacketIdent]*packetHistoryRecord),
	}

	s.wg.Add(2)
	go s.resendUnAckedPackets()
	go s.resendPendingRequests()

	return s
}

func (s *sendManager) resendUnAckedPackets() {
	defer s.wg.Done()

	t := time.NewTicker(PACKET_RESEND)

LOOP:
	for {
		select {
		case <-s.ctx.Done():
			break LOOP
		case <-t.C:
			s.mu.Lock()

			now := time.Now()
			resendTime := now.Add(-PACKET_RESEND)

			for _, r := range s.history {

				if r.updated.Before(resendTime) {
					slog.Info(fmt.Sprintf("[CLIENT SEND MANAGER] Resending packet with Id %v", r.packet.Header.MessageId))
					err := s.sendWithoutSet(r.conn, r.addr, r.packet)
					if err != nil {
						slog.Error(err.Error())
						continue
					}
					r.updated = now
				}
			}
			s.mu.Unlock()

		default:
			continue
		}
	}
}

func (s *sendManager) resendPendingRequests() {
	defer s.wg.Done()

	t := time.NewTicker(PACKET_RESEND)

LOOP:
	for {
		select {
		case <-s.ctx.Done():
			break LOOP
		case <-t.C:
			s.mu.Lock()

			now := time.Now()
			expireTime := now.Add(-PACKET_TTL)
			resendTime := now.Add(-PACKET_RESEND)

			for k, r := range s.requestHistory {
				if r.created.Before(expireTime) {
					delete(s.requestHistory, k)
					continue
				}

				if r.updated.Before(resendTime) {
					slog.Info(fmt.Sprintf("[CLIENT SEND MANAGER] Resending request with Id %v", r.packet.Header.MessageId))
					err := s.sendWithoutSet(r.conn, r.addr, r.packet)
					if err != nil {
						slog.Error(err.Error())
						continue
					}
					r.updated = now
				}
			}
			s.mu.Unlock()

		default:
			continue
		}
	}
}

func (s *sendManager) sendWithoutSet(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) error {

	b, err := p.MarshalBinary()
	if err != nil {
		return err
	}

	if _, err := c.WriteToUDP(b, a); err != nil {
		return err
	}

	return nil
}

func (s *sendManager) send(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) error {

	if err := s.sendWithoutSet(c, a, p); err != nil {
		return err
	}
	s.set(c, a, p)

	return nil
}

func (s *sendManager) set(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r, exists := s.history[protocol.ExtractIdentFromPacket(p)]; exists {
		r.updated = time.Now()
	} else {
		s.history[protocol.ExtractIdentFromPacket(p)] = newPacketHistoryRecord(c, a, p)
	}

	if p.Header.MessageType == proto_defs.MessageTypeRequest {
		if r, exists := s.requestHistory[protocol.ExtractIdentFromPacket(p)]; exists {
			r.updated = time.Now()
		} else {
			s.requestHistory[protocol.ExtractIdentFromPacket(p)] = newPacketHistoryRecord(c, a, p)
		}
	}
}

func (s *sendManager) clear(i *protocol.PacketIdent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.history, *i)
}

func (s *sendManager) clearRequest(i *protocol.PacketIdent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.requestHistory, *i)
}

func (s *sendManager) get(i *protocol.PacketIdent) (*protocol.Packet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, exists := s.history[*i]; exists {
		return p.packet, nil
	} else {
		return nil, errors.New("packet does not exist")
	}
}

func (s *sendManager) close() {
	s.cancelFunc()
	// Wait for background functions to finish
	s.wg.Wait()
}
