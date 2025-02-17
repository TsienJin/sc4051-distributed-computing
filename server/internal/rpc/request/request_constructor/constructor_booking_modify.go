package request_constructor

import (
	"server/internal/interfaces"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"server/internal/rpc/request"
)

func NewBookingModifyPacket(
	id uint16,
	deltaHours int,
) interfaces.RpcRequestConstructor {
	return func() ([]*protocol.Packet, error) {
		payload := request.NewBookingModifyPayload(id, deltaHours)
		payloadBytes, err := payload.MarshalBinary()
		if err != nil {
			return nil, err
		}

		r := request.Request{
			MethodIdentifier: request.MethodIdentifierBookingUpdate,
			Payload:          payloadBytes,
		}

		headerDistilled := &protocol.PacketHeaderDistilled{
			Version:     proto_defs.ProtocolV1,
			MessageId:   proto_defs.NewMessageId(),
			MessageType: proto_defs.MessageTypeRequest,
			RequireAck:  true,
		}

		message, err := protocol.NewMessage(headerDistilled, &r)
		if err != nil {
			return nil, err
		}

		return message.ToPackets()
	}
}
