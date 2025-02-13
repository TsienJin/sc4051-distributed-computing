package request

import "server/internal/bookings"

type FacilityCreatePayload struct {
	Name bookings.FacilityName
}

func (f *FacilityCreatePayload) MarshalBinary() ([]byte, error) {
	return []byte(f.Name), nil
}

func (f *FacilityCreatePayload) UnmarshalBinary(data []byte) error {
	f.Name = bookings.FacilityName(data)
	return nil
}
