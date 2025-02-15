package request_constructor

import (
	"server/internal/bookings"
	"server/internal/interfaces"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"server/internal/rpc/request"
)

func NewFacilityCreatePacket(name string) interfaces.RpcRequestConstructor {

	return func() ([]*protocol.Packet, error) {
		payload := &request.FacilityCreatePayload{
			Name: bookings.FacilityName(name),
		}

		payloadByte, err := payload.MarshalBinary()
		if err != nil {
			return nil, err
		}

		r := request.Request{
			MethodIdentifier: request.MethodIdentifierFacilityCreate,
			Payload:          payloadByte,
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
