package request

import "server/internal/bookings"

type FacilityQueryPayload struct {
	Name bookings.FacilityName
	Days int
}

func (f *FacilityQueryPayload) UnmarshalBinary(data []byte) error {
	f.Days = int(data[0])
	f.Name = bookings.FacilityName(data[1:])
	return nil
}
