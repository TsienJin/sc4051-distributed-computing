package request

import (
	"encoding/binary"
	"server/internal/bookings"
	"time"
)

type BookingMakePayload struct {
	Name  bookings.FacilityName
	Start time.Time
	End   time.Time
}

func (b *BookingMakePayload) UnmarshalBinary(data []byte) error {

	b.Name = bookings.FacilityName(data[7:])

	startHourOffset := int(binary.BigEndian.Uint32(append([]byte{0x00}, data[0:4]...)))
	endHourOffset := int(binary.BigEndian.Uint32(append([]byte{0x00}, data[4:7]...)))

	unixTime := time.Unix(0, 0)

	b.Start = unixTime.Add(time.Duration(startHourOffset) * time.Hour)
	b.End = unixTime.Add(time.Duration(endHourOffset) * time.Hour)

	return nil
}
