package request

import (
	"bytes"
	"encoding/binary"
)

type BookingDeletePayload struct {
	Id uint16
}

func NewBookingDeletePayload(id uint16) *BookingDeletePayload {
	return &BookingDeletePayload{
		Id: id,
	}
}

func (b *BookingDeletePayload) UnmarshalBinary(data []byte) error {
	b.Id = binary.BigEndian.Uint16(data)
	return nil
}

func (b *BookingDeletePayload) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, b.Id); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
