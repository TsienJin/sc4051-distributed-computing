package handle

import (
	"bytes"
	"log/slog"
	"math/bits"
	"net"
	"server/internal/protocol"
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

func (m *MessagePartial) IsComplete() (*protocol.Message, bool) {
	m.RLock()
	defer m.RUnlock()

	received := 0
	for _, b := range m.Bitmap {
		received += bits.OnesCount8(b)
	}

	if received == m.Total {
		slog.Debug("MessagePartial complete!", "MessageId", m.DistilledHeader.MessageId)
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
func (m *MessagePartial) UpsertPacket(p *protocol.Packet) {
	byteIdx, mask := m.GetPacketBitmapPosition(p)

	m.Lock()
	defer m.Unlock()

	if m.Bitmap[byteIdx]&mask == 1 {
		slog.Info("Packet already added to partial", "MessageId", p.Header.MessageId, "PacketNumber", p.Header.PacketNumber)
		return
	}

	if m.DistilledHeader == nil {
		m.DistilledHeader = p.Header.ToDistilled()
	}

	// Add payload to partial
	m.Bitmap[byteIdx] |= mask
	m.Payloads[p.Header.PacketNumber] = p.Payload
	m.LastUpdated = time.Now()
	slog.Info("Added new packet to partial message", "MessageId", m.DistilledHeader.MessageId)
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
		messageAssembler = &MessageAssembler{}
	})

	return messageAssembler
}

func (m *MessageAssembler) AssembleMessageFromPacket(c *net.UDPConn, a *net.UDPAddr, p *protocol.Packet) {
	m.Lock()
	defer m.Unlock()

	// Extract packet identifier for the supposed message
	ident := protocol.ExtractIdentFromPacket(p)

	// Check if the message has already been completed (prevents duplicate messages)
	if _, exists := m.Complete[ident]; exists {
		slog.Info("Message has already been assembled and handed off", "MessageId", p.Header.MessageId)
		return
	}

	// Add packet to MessagePartial
	if mp, exists := m.Incomplete[ident]; exists {
		mp.UpsertPacket(p)
	} else {
		m.Incomplete[ident] = NewMessagePartial(c, a, int(p.Header.TotalPackets))
		m.Incomplete[ident].UpsertPacket(p)
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
