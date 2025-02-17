package request

import (
	"encoding/binary"
	"fmt"
	"server/internal/bookings"
)

type FacilityMonitorPayload struct {
	Name bookings.FacilityName
	Ttl  int
}

func (f *FacilityMonitorPayload) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("data length too short: %d, expected at least 3", len(data))
	}

	// Extract TTL from the first 3 bytes. We promote it to a uint32 for `binary.BigEndian.Uint32`, then cast it to int.
	ttlBytes := []byte{0x00, data[0], data[1], data[2]} // Prepend with 0x00
	f.Ttl = int(binary.BigEndian.Uint32(ttlBytes))
	f.Name = bookings.FacilityName(data[3:])

	return nil
}

func (f *FacilityMonitorPayload) MarshalBinary() ([]byte, error) {
	nameBytes := []byte(f.Name)
	data := make([]byte, 3+len(nameBytes)) // 3 bytes for TTL, rest for name

	// Convert int TTL to bytes.  We take the last 3 bytes of the uint32 representation
	ttlUint32 := uint32(f.Ttl)
	data[0] = byte(ttlUint32 >> 16) // Most significant byte
	data[1] = byte(ttlUint32 >> 8)
	data[2] = byte(ttlUint32) // Least significant byte

	// Copy the name.
	copy(data[3:], nameBytes)

	return data, nil
}
