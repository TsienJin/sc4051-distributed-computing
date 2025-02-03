package request

import "server/internal/bookings"

type FacilityCreatePayload struct {
	Name bookings.FacilityName
}

func (f *FacilityCreatePayload) UnmarshalBinary(data []byte) error {
	f.Name = bookings.FacilityName(data)
	return nil
}
