package request

import (
	"encoding/binary"
	"fmt"
)

type flags uint8

const (
	reverseTime flags = 1 << iota
)

type BookingModifyPayload struct {
	Id        uint16
	DeltaHour int
}

func NewBookingModifyPayload(id uint16, deltaHour int) *BookingModifyPayload {
	return &BookingModifyPayload{
		Id:        id,
		DeltaHour: deltaHour,
	}
}

func (b *BookingModifyPayload) UnmarshalBinary(data []byte) error {

	if len(data) != 6 {
		return fmt.Errorf("payload for BookingModifyPayload must be 6 bytes, received: %d", len(data))
	}

	b.Id = binary.BigEndian.Uint16(data[:2])

	multiplier := 1
	f := data[2]
	if f&0x01 == 1 {
		multiplier = -1
	}

	b.DeltaHour = multiplier * int(binary.BigEndian.Uint32(append([]byte{0x00}, data[3:]...)))

	return nil

}

func (b *BookingModifyPayload) MarshalBinary() ([]byte, error) {

	buffer := make([]byte, 6) // 2 bytes for Id, 1 byte for flag, 6 bytes for DeltaHour

	// Encode Id (2 bytes, BigEndian)
	binary.BigEndian.PutUint16(buffer[0:2], b.Id)

	// Determine the multiplier flag
	var flag byte
	absDeltaHour := b.DeltaHour
	if b.DeltaHour < 0 {
		flag = 0x01
		absDeltaHour = -b.DeltaHour
	}

	// Store the flag in the 3rd byte
	buffer[2] = flag

	// Encode DeltaHour as 4 bytes, ensuring correct format with padding
	deltaBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(deltaBytes, uint32(absDeltaHour))

	// Copy last 3 bytes of deltaBytes (skip first padding byte)
	copy(buffer[3:], deltaBytes[1:4])

	return buffer, nil
}
