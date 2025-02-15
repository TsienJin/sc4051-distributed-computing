package request

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"server/internal/bookings"
	"testing"
	"time"
)

func TestBookingMakePayload_FullCycle(t *testing.T) {

	name := "TestBookingMakePayload_MarshalBinary"

	now := time.Now()
	startHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())

	end := now.Add(time.Duration(2) * time.Hour)
	endHour := time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), 0, 0, 0, end.Location())

	original := NewBookingMakePayload(name, now, end)

	if !original.Start.Equal(startHour) {
		t.Logf("E: %v", startHour)
		t.Logf("R: %v", original.Start)
		t.Error("Start time does not match correct hour")
	}

	if !original.End.Equal(endHour) {
		t.Logf("E: %v", endHour)
		t.Logf("R: %v", original.End)
		t.Error("End time does not match correct hour")
	}

	binary, err := original.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	var reconstructed BookingMakePayload
	if err := reconstructed.UnmarshalBinary(binary); err != nil {
		t.Error(err)
	}

	if !cmp.Equal(*original, reconstructed) {
		t.Logf("E: %v", reconstructed)
		t.Logf("R: %v", *original)
		t.Error("Original and reconstructed BookingMakePayload's do not match")
	}

}

func TestBookingMakePayload_UnmarshalBinary(t *testing.T) {

	name := "TestBookingMakePayload_UnmarshalBinary"

	startTime := time.Now()
	start := uint32(startTime.Unix() / 3600)
	end := uint32(startTime.Add(time.Duration(2)*time.Hour).Unix() / 3600)

	buf := new(bytes.Buffer)

	buf.WriteByte(byte(start >> 16))
	buf.WriteByte(byte(start >> 8))
	buf.WriteByte(byte(start))

	buf.WriteByte(byte(end >> 16))
	buf.WriteByte(byte(end >> 8))
	buf.WriteByte(byte(end))

	buf.Write([]byte(name))

	var bookingMakePayload BookingMakePayload
	if err := bookingMakePayload.UnmarshalBinary(buf.Bytes()); err != nil {
		t.Error(err)
	}
}

func TestBookingMakePayload_GetBooking(t *testing.T) {

	name := "TestBookingMakePayload_GetBooking"

	manager := bookings.GetManager()
	if err := manager.NewFacility(bookings.FacilityName(name)); err != nil {
		t.Error(err)
	}

	startTimeFull := time.Now()
	startTime := time.Date(startTimeFull.Year(), startTimeFull.Month(), startTimeFull.Day(), startTimeFull.Hour(), 0, 0, 0, startTimeFull.Location())
	start := uint32(startTime.Unix() / 3600)
	end := uint32(startTime.Add(time.Duration(2)*time.Hour).Unix() / 3600)

	buf := new(bytes.Buffer)

	buf.WriteByte(byte(start >> 16))
	buf.WriteByte(byte(start >> 8))
	buf.WriteByte(byte(start))

	buf.WriteByte(byte(end >> 16))
	buf.WriteByte(byte(end >> 8))
	buf.WriteByte(byte(end))

	buf.Write([]byte(name))

	var bookingMakePayload BookingMakePayload
	if err := bookingMakePayload.UnmarshalBinary(buf.Bytes()); err != nil {
		t.Error(err)
	}

	booking, err := bookingMakePayload.GetBooking()
	if err != nil {
		t.Error(err)
	}

	if booking.Start != startTime {
		t.Error("Booking start time doesn't match expected time")
	}

}

func TestBookingMakePayload_GetBookingTimTestCase(t *testing.T) {

	name := "TestBookingMakePayload_GetBookingTimTestCase"

	manager := bookings.GetManager()
	if err := manager.NewFacility(bookings.FacilityName(name)); err != nil {
		t.Error(err)
	}

	start := uint32(483217)
	end := uint32(483219)

	buf := new(bytes.Buffer)

	buf.WriteByte(byte(start >> 16))
	buf.WriteByte(byte(start >> 8))
	buf.WriteByte(byte(start))

	buf.WriteByte(byte(end >> 16))
	buf.WriteByte(byte(end >> 8))
	buf.WriteByte(byte(end))

	buf.Write([]byte(name))

	var bookingMakePayload BookingMakePayload
	if err := bookingMakePayload.UnmarshalBinary(buf.Bytes()); err != nil {
		t.Error(err)
	}

	_, err := bookingMakePayload.GetBooking()
	if err != nil {
		t.Error(err)
	}

}
