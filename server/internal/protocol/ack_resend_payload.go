package protocol

import (
	"fmt"
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
	if len(data) < 17 {
		return fmt.Errorf("AckResendPayload too short to be valid: % X", data)
	}

	a.Id = proto_defs.MessageId(data[0:16])
	a.PacketNumber = data[16]
	return nil
}

func (a *AckResendPayload) ToPacketIdent() *PacketIdent {
	return &PacketIdent{
		MessageId:    a.Id,
		PacketNumber: a.PacketNumber,
	}
}
