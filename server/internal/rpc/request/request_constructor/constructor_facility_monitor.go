package request_constructor

import (
	"server/internal/interfaces"
	"server/internal/protocol"
	"server/internal/protocol/proto_defs"
	"server/internal/rpc/request"
)

func NewFacilityMonitorPacket(name string, ttl int) interfaces.RpcRequestConstructor {
	return func() ([]*protocol.Packet, error) {
		payload := request.NewFacilityMonitorPayload(name, ttl)
		payloadByte, err := payload.MarshalBinary()
		if err != nil {
			return nil, err
		}

		r := request.Request{
			MethodIdentifier: request.MethodIdentifierFacilityMonitor,
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
