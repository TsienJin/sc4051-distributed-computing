package request

import (
	"bytes"
	"server/internal/bookings"
)

type FacilityDeletePayload struct {
	Name bookings.FacilityName
}

func NewFacilityDeletePayload(name string) *FacilityDeletePayload {
	return &FacilityDeletePayload{Name: bookings.FacilityName(name)}
}

func (f *FacilityDeletePayload) UnmarshalBinary(data []byte) error {
	f.Name = bookings.FacilityName(data)
	return nil
}

func (f *FacilityDeletePayload) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	buf.Write([]byte(f.Name))
	return buf.Bytes(), nil
}
