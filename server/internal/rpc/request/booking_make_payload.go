package request

import (
	"bytes"
	"encoding/binary"
	"server/internal/bookings"
	"time"
)

type BookingMakePayload struct {
	Name  bookings.FacilityName
	Start time.Time
	End   time.Time
}

func NewBookingMakePayload(
	name string,
	start time.Time,
	end time.Time,
) *BookingMakePayload {

	startHour := time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, start.Location())
	endHour := time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), 0, 0, 0, end.Location())

	return &BookingMakePayload{
		Name:  bookings.FacilityName(name),
		Start: startHour,
		End:   endHour,
	}
}

func (b *BookingMakePayload) MarshalBinary() ([]byte, error) {

	startTimeFull := b.Start
	startTime := time.Date(startTimeFull.Year(), startTimeFull.Month(), startTimeFull.Day(), startTimeFull.Hour(), 0, 0, 0, startTimeFull.Location())
	start := uint32(startTime.Unix() / 3600)

	endTimeFull := b.End
	endTime := time.Date(endTimeFull.Year(), endTimeFull.Month(), endTimeFull.Day(), endTimeFull.Hour(), 0, 0, 0, endTimeFull.Location())
	end := uint32(endTime.Unix() / 3600)

	buf := new(bytes.Buffer)

	buf.WriteByte(byte(start >> 16))
	buf.WriteByte(byte(start >> 8))
	buf.WriteByte(byte(start))

	buf.WriteByte(byte(end >> 16))
	buf.WriteByte(byte(end >> 8))
	buf.WriteByte(byte(end))

	buf.Write([]byte(b.Name))

	return buf.Bytes(), nil
}

func (b *BookingMakePayload) UnmarshalBinary(data []byte) error {

	b.Name = bookings.FacilityName(data[6:])

	startHourOffset := int(binary.BigEndian.Uint32(append([]byte{0x00}, data[0:3]...)))
	endHourOffset := int(binary.BigEndian.Uint32(append([]byte{0x00}, data[3:6]...)))

	unixTime := time.Unix(0, 0)

	b.Start = unixTime.Add(time.Duration(startHourOffset) * time.Hour)
	b.End = unixTime.Add(time.Duration(endHourOffset) * time.Hour)

	return nil
}

func (b *BookingMakePayload) GetBooking() (bookings.Booking, error) {
	return bookings.NewBooking(
		bookings.BookingWithRandomId(),
		bookings.BookingWithStartTime(b.Start),
		bookings.BookingWithEndTime(b.End),
	)
}
