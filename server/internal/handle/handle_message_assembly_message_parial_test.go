package handle

import (
	"bytes"
	"net"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"testing"
)

func TestMessagePartial_IsComplete(t *testing.T) {

	var mockConn *net.UDPConn = nil
	var mockAddr *net.UDPAddr = nil

	messageId := proto_defs.NewMessageId()
	totalPackets := 10 // This will require 2 bytes

	packets := make([]*protocol.Packet, totalPackets)

	payloadBytesAll := make([]byte, totalPackets)

	for i := 0; i < totalPackets; i++ {
		payloadBytesAll[i] = uint8(i)

		packetHeader, _ := protocol.NewPacketHeader(
			protocol.PacketHeaderWithMessageType(proto_defs.MessageTypeRequest),
			protocol.PacketHeaderWithVersion(proto_defs.ProtocolV1),
			protocol.PacketHeaderWithMessageId(messageId),
			protocol.PacketHeaderWithTotalPackets(uint8(totalPackets)),
			protocol.PacketHeaderWithPacketNumber(uint8(i)),
		)
		packet, _ := protocol.NewPacket(*packetHeader, []byte{payloadBytesAll[i]})
		packets[i] = packet
	}

	m := NewMessagePartial(mockConn, mockAddr, totalPackets)

	for _, p := range packets {
		m.UpsertPacket(p)
	}

	message, complete := m.IsComplete()
	if !complete {
		t.Error("Message supposed to be complete")
	}

	if !bytes.Equal(message.Payload, payloadBytesAll) {
		t.Error("Bytes are supposed to match")
	}

}

func TestMessagePartial_UpsertPacket(t *testing.T) {

	var mockConn *net.UDPConn = nil
	var mockAddr *net.UDPAddr = nil

	messageId := proto_defs.NewMessageId()
	totalPackets := 10 // This will require 2 bytes

	packets := make([]*protocol.Packet, totalPackets)

	for i := 0; i < totalPackets; i++ {
		packetHeader, _ := protocol.NewPacketHeader(
			protocol.PacketHeaderWithMessageType(proto_defs.MessageTypeRequest),
			protocol.PacketHeaderWithVersion(proto_defs.ProtocolV1),
			protocol.PacketHeaderWithMessageId(messageId),
			protocol.PacketHeaderWithTotalPackets(uint8(totalPackets)),
			protocol.PacketHeaderWithPacketNumber(uint8(i)),
		)
		packet, _ := protocol.NewPacket(*packetHeader, []byte{})
		packets[i] = packet
	}

	m := NewMessagePartial(mockConn, mockAddr, totalPackets)

	m.UpsertPacket(packets[0])

	if !bytes.Equal(m.Bitmap, []byte{0x01, 0x00}) {
		t.Error("Upsert did not update the correct bit!")
	}

	m.UpsertPacket(packets[1])

	if !bytes.Equal(m.Bitmap, []byte{0x03, 0x00}) {
		t.Logf("% X", m.Bitmap)
		t.Error("Upsert did not update the correct bit!")
	}

	m.UpsertPacket(packets[2])

	if !bytes.Equal(m.Bitmap, []byte{0x07, 0x00}) {
		t.Logf("% X", m.Bitmap)
		t.Error("Upsert did not update the correct bit!")
	}

	for _, p := range packets {
		m.UpsertPacket(p)
	}

	if !bytes.Equal(m.Bitmap, []byte{0xFF, 0x03}) {
		t.Logf("% X", m.Bitmap)
		t.Error("Upsert did not update the correct bit!")
	}

}
