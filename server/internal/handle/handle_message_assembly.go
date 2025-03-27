package handle

import (
	"bytes"
	"errors"
	"log/slog"
	"math/bits"
	"net"
	"server/internal/network"
	"server/internal/protocol"
	"server/internal/protocol/constructors"
	"server/internal/rpc/response"
	"server/internal/vars"
	"sync"
	"time"
)

type MessagePartial struct {
	sync.RWMutex

	DistilledHeader *protocol.PacketHeaderDistilled

	Conn        *net.UDPConn
	Addr        *net.UDPAddr
	Bitmap      []byte
	Payloads    [][]byte
	Total       int
	LastUpdated time.Time
}

func NewMessagePartial(
	conn *net.UDPConn,
	addr *net.UDPAddr,
	nPackets int,
) *MessagePartial {
	nBytesForBitmap := (nPackets + 7) / 8
	return &MessagePartial{
		Conn:        conn,
		Addr:        addr,
		Bitmap:      make([]byte, nBytesForBitmap),
		Payloads:    make([][]byte, nPackets),
		Total:       nPackets,
		LastUpdated: time.Now(),
	}
}

func (m *MessagePartial) GetMissingPackets() []uint8 {
	m.RLock()
	defer m.RUnlock()

	// Determine missing packets
	var missing []uint8

	for i := 0; i < m.Total; i++ {
		byteIdx := i / 8
		bitMask := byte(1 << (i % 8))

		if m.Bitmap[byteIdx]&bitMask == 0 {
			// Packet missing
			slog.Info("Missing packet", "PacketNumber", i, "MessageId", m.DistilledHeader.MessageId)
			missing = append(missing, uint8(i))
		}
	}

	return missing
}

func (m *MessagePartial) RequestMissingPackets() {
	m.RLock()
	defer m.RUnlock()

	if time.Now().Before(m.LastUpdated.Add(time.Duration(1) * time.Second)) {
		return
	}

	if m.IsCompleteCheck() {
		slog.Warn("Partial message is complete was but RequestMissingPackets() was called", "MessageId", m.DistilledHeader.MessageId)
		return
	}

	missingIds := m.GetMissingPackets()
	if len(missingIds) == 0 {
		slog.Warn("No missing packets, but RequestMissingPackets() was called", "MessageId", m.DistilledHeader.MessageId)
	}

	slog.Info("Requesting for missing packets", "PacketIds", missingIds, "MessageId", m.DistilledHeader.MessageId)

	for _, i := range missingIds {
		p, err := constructors.NewRequestResend(m.DistilledHeader.MessageId, i)
		if err != nil {
			slog.Error("Unable to create Request Resend packet", "err", err)
			continue
		}
		if err := network.SendPacket(m.Conn, m.Addr, p); err != nil {
			slog.Error("Unable to send Request Resend packet", "err", err)
		}
	}

}

func (m *MessagePartial) IsCompleteCheck() bool {
	m.RLock()
	defer m.RUnlock()

	received := 0
	for _, b := range m.Bitmap {
		received += bits.OnesCount8(b)
	}

	return received == m.Total
}

func (m *MessagePartial) IsComplete() (*protocol.Message, bool) {
	m.RLock()
	defer m.RUnlock()

	if m.IsCompleteCheck() {
		slog.Info("MessagePartial complete!", "MessageId", m.DistilledHeader.MessageId)
		return protocol.NewMessageFromBytes(m.DistilledHeader, bytes.Join(m.Payloads, nil)), true
	}

	return nil, false
}

// GetPacketBitmapPosition
// Returns
// - Byte index, int
// - bitmask OHE, byte
func (m *MessagePartial) GetPacketBitmapPosition(p *protocol.Packet) (int, byte) {
	byteIdx := (int(p.Header.PacketNumber)) / 8
	bitMask := byte(1 << (p.Header.PacketNumber % 8))
	return byteIdx, bitMask
}

// UpsertPacket checks if the packet has already been added to MessagePartial.
func (m *MessagePartial) UpsertPacket(p *protocol.Packet) error {

	// Ensure that the packet number does not exceed range
	if p.Header.PacketNumber >= uint8(m.Total) {
		slog.Error("Packet number does exceeds total number of packets", "PacketNumber", p.Header.PacketNumber, "Total", m.Total)
		return errors.New("packet number exceeds total number of packets")
	}

	byteIdx, mask := m.GetPacketBitmapPosition(p)

	m.Lock()
	defer m.Unlock()

	if m.Bitmap[byteIdx]&mask == 1 {
		slog.Info("Packet already added to partial", "MessageId", p.Header.MessageId, "PacketNumber", p.Header.PacketNumber)
		return nil
	}

	if m.DistilledHeader == nil {
		m.DistilledHeader = p.Header.ToDistilled()
	}

	// Add payload to partial
	m.Bitmap[byteIdx] |= mask
	m.Payloads[p.Header.PacketNumber] = p.Payload
	m.LastUpdated = time.Now()
	slog.Info("Added new packet to partial message", "MessageId", m.DistilledHeader.MessageId)
	return nil
}

type MessageAssembler struct {
	sync.RWMutex
	Incomplete map[protocol.PacketIdent]*MessagePartial
	Complete   map[protocol.PacketIdent]struct{}
}

var onceMessageAssembler sync.Once
var messageAssembler *MessageAssembler

func GetMessageAssembler() *MessageAssembler {
	onceMessageAssembler.Do(func() {
		messageAssembler = &MessageAssembler{
			Incomplete: make(map[protocol.PacketIdent]*MessagePartial),
			Complete:   make(map[protocol.PacketIdent]struct{}),
		}

		// Create an interval to request missing packets on existing incomplete packets
		t := time.NewTicker(time.Duration(vars.GetStaticEnv().MessageAssemblerIntervals) * time.Millisecond)
		go func() {
			defer t.Stop()
			for range t.C {
				messageAssembler.RequestMissingPackets()
			}
		}()
	})

	return messageAssembler
}

func (m *MessageAssembler) RequestMissingPackets() {
	m.RLock()
	defer m.RUnlock()

	if len(m.Incomplete) == 0 {
		return
	}

	slog.Debug("Requesting missing packets in all message partials")
	for _, partial := range m.Incomplete {
		go partial.RequestMissingPackets()
	}
	slog.Debug("Requesting missing packets done")
}

func (m *MessageAssembler) AssembleMessageFromPacket(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) {
	m.Lock()
	defer m.Unlock()

	// Extract packet identifier for the supposed message
	ident := protocol.ExtractIdentFromPacket(p)

	// Check if the message has already been completed (prevents duplicate messages)
	if _, exists := m.Complete[ident]; exists && vars.GetStaticEnv().EnableDuplicateFiltering {
		slog.Info("Message has already been assembled and handed off, resending cached response", "MessageId", p.Header.MessageId)
		res, err := response.GetResponseHistoryInstance().GetResponse(p.Header.MessageId)
		if err != nil {
			slog.Error("Unable to resend cached response", "err", err)
		}
		if res == nil {
			slog.Warn("Response has yet to be completed, dropping request packet")
			return
		}
		response.SendResponse(c, a, res)
		return
	}

	// Add packet to MessagePartial
	if mp, exists := m.Incomplete[ident]; exists {
		slog.Info("Upsert packet that already exists", "PacketIdent", ident)
		if err := mp.UpsertPacket(p); err != nil {
			return
		}
	} else {
		slog.Info("Setting new partial", "PacketIdent", ident)
		m.Incomplete[ident] = NewMessagePartial(c, a, int(p.Header.TotalPackets))
		if err := m.Incomplete[ident].UpsertPacket(p); err != nil {
			return
		}
	}

	if message, completed := m.Incomplete[ident].IsComplete(); completed {
		slog.Info("Message completed", "MessageId", ident.MessageId)

		// Shift record to be completed
		delete(m.Incomplete, ident)
		m.Complete[ident] = struct{}{}

		// Handoff message to be processed
		go IncomingMessage(c, a, message)
	}

}

func AssembleMessageFromPacket(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) {
	GetMessageAssembler().AssembleMessageFromPacket(c, a, p)
}
