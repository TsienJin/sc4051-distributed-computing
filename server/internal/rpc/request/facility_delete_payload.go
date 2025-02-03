package request

import "server/internal/bookings"

type FacilityDeletePayload struct {
	Name bookings.FacilityName
}

func (f *FacilityDeletePayload) UnmarshalBinary(data []byte) error {
	f.Name = bookings.FacilityName(data)
	return nil
}
