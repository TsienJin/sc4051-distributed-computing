package request

import "server/internal/bookings"

type BookingDeletePayload struct {
	Name bookings.FacilityName
}

func (b *BookingDeletePayload) UnmarshalBinary(data []byte) error {
	b.Name = bookings.FacilityName(data)
	return nil
}
