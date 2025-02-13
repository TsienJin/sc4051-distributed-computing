package constructors

import (
	"errors"
	"log/slog"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
)

// NewAck creates a packet to send as an acknowledgement packet.
func NewAck(
	originalId proto_defs.MessageId,
	packetNumber uint8,
) (*protocol.Packet, error) {

	ack := protocol.AckResendPayload{
		Id:           originalId,
		PacketNumber: packetNumber,
	}
	payload, err := ack.MarshalBinary()
	if err != nil {
		slog.Error("Unable to marshal ack payload into binary", "AckPayload", ack)
		return nil, errors.New("unable to marshal ack payload into binary")
	}

	h, err := protocol.NewPacketHeader(
		protocol.PacketHeaderWithVersion(proto_defs.ProtocolV1),
		protocol.PacketHeaderWithMessageId(proto_defs.NewMessageId()),
		protocol.PacketHeaderWithMessageType(proto_defs.MessageTypeAcknowledge),
		protocol.PacketHeaderWithPacketNumber(0),
		protocol.PacketHeaderWithTotalPackets(1),
		protocol.PacketHeaderWithPayloadLength(uint16(len(payload))),
	)
	if err != nil {
		slog.Error("Unable to generate packet header for ack")
		return nil, errors.New("unable to generate packet header for ack")
	}

	p, err := protocol.NewPacket(
		*h,
		payload,
	)

	return p, nil
}
