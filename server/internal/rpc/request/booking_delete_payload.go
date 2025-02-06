package request

import (
	"encoding/binary"
)

type BookingDeletePayload struct {
	Id uint16
}

func (b *BookingDeletePayload) UnmarshalBinary(data []byte) error {
	b.Id = binary.BigEndian.Uint16(data)
	return nil
}
