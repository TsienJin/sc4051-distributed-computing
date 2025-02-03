package request

import (
	"encoding/binary"
	"server/internal/bookings"
)

type FacilityMonitorPayload struct {
	Name bookings.FacilityName
	Ttl  int
}

func (f *FacilityMonitorPayload) UnmarshalBinary(data []byte) error {
	tempByteArr := append([]byte{0x00}, data[0:4]...)
	f.Ttl = int(binary.BigEndian.Uint32(tempByteArr))
	f.Name = bookings.FacilityName(data[4:])
	return nil
}
