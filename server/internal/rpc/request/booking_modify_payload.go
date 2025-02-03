package request

import "encoding/binary"

type flags uint8

const (
	reverseTime flags = 1 << iota
)

type BookingModifyPayload struct {
	Id        uint16
	DeltaHour int
}

func (b *BookingModifyPayload) UnmarshalBinary(data []byte) error {

	b.Id = binary.BigEndian.Uint16(data[:2])

	multiplier := 1
	f := data[2]
	if f&0x01 == 1 {
		multiplier = -1
	}

	b.DeltaHour = multiplier * int(binary.BigEndian.Uint32(append([]byte{0x00}, data[3:]...)))

	return nil

}
