package protocol

import (
	"server/internal/protocol/proto_defs"
)

// AckResendPayload is the payload layout for Acknowledgements and Resend packets
type AckResendPayload struct {
	Id           proto_defs.MessageId
	PacketNumber uint8
}

func (a *AckResendPayload) MarshalBinary() ([]byte, error) {
	return append(a.Id[:], a.PacketNumber), nil
}

func (a *AckResendPayload) UnmarshalBinary(data []byte) error {
	a.Id = proto_defs.MessageId(data[0:4])
	a.PacketNumber = data[4]
	return nil
}

func (a *AckResendPayload) ToPacketIdent() *PacketIdent {
	return &PacketIdent{
		MessageId:    a.Id,
		PacketNumber: a.PacketNumber,
	}
}
