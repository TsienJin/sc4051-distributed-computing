package bookings

import (
	"bytes"
	"testing"
	"time"
)

func TestFacility_clean(t *testing.T) {

	timeNow := time.Now()

	old := &Booking{
		Id:    1,
		Start: timeNow.Add(time.Duration(-5) * time.Hour),
		End:   timeNow.Add(time.Duration(-4) * time.Hour),
	}

	current := &Booking{
		Id:    1,
		Start: timeNow.Add(time.Duration(-1) * time.Hour),
		End:   timeNow.Add(time.Duration(2) * time.Hour),
	}

	future := &Booking{
		Id:    1,
		Start: timeNow.Add(time.Duration(5) * time.Hour),
		End:   timeNow.Add(time.Duration(7) * time.Hour),
	}

	f := NewFacility(FacilityName("Testing"))

	f.Bookings = []*Booking{old}
	f.clean()
	if len(f.Bookings) != 0 {
		t.Error("clean() failed to clean up old booking")
	}

	f.Bookings = []*Booking{current, future}
	f.clean()
	if len(f.Bookings) != 2 {
		t.Error("clean() cleaned up a bit too much")
	}

}

func TestFacility_QueryAvailability(t *testing.T) {
	currentTime := time.Now()
	tmr := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day()+1, 0, 0, 0, 0, time.Local)

	b1 := &Booking{
		Id:    1,
		Start: tmr.Add(time.Duration(0) * time.Hour),
		End:   tmr.Add(time.Duration(3) * time.Hour),
	}

	f := NewFacility(FacilityName("Testing"))

	_ = f.Book(*b1)

	expectedByteArr := []byte{0x00, 0x00, 0x00, 0xE0, 0x00, 0x00}

	if !bytes.Equal(f.QueryAvailability(2), expectedByteArr) {
		t.Logf("E: % X", expectedByteArr)
		t.Logf("R: % X", f.QueryAvailability(2))
		t.Error("Availability does not match expected")
	}

}

func TestFacility_Book(t *testing.T) {

	currentTime := time.Now()

	b1 := &Booking{
		Id:    1,
		Start: currentTime.Add(time.Duration(1) * time.Hour),
		End:   currentTime.Add(time.Duration(3) * time.Hour),
	}
	b2 := &Booking{
		Id:    2,
		Start: currentTime.Add(time.Duration(3) * time.Hour),
		End:   currentTime.Add(time.Duration(4) * time.Hour),
	}
	b3 := &Booking{
		Id:    3,
		Start: currentTime.Add(time.Duration(1) * time.Hour),
		End:   currentTime.Add(time.Duration(2) * time.Hour),
	}

	f := NewFacility(FacilityName("Testing"))

	if err := f.Book(*b1); err != nil {
		t.Error("Unexpected booking error")
	}
	if len(f.Bookings) != 1 {
		t.Error("Booking was not inserted properly")
	}

	if err := f.Book(*b2); err != nil {
		t.Error("Unexpected booking error")
	}
	if len(f.Bookings) != 2 {
		t.Error("Booking was not inserted properly")
	}

	if err := f.Book(*b3); err == nil {
		t.Error("Expected booking clash error, but none were raised")
	}
	if len(f.Bookings) != 2 {
		t.Error("Booking was not inserted properly")
	}

}

func TestFacility_UpdateBooking(t *testing.T) {

	currentTime := time.Now()

	b1 := &Booking{
		Id:    1,
		Start: currentTime.Add(time.Duration(1) * time.Hour),
		End:   currentTime.Add(time.Duration(3) * time.Hour),
	}
	b2 := &Booking{
		Id:    2,
		Start: currentTime.Add(time.Duration(3) * time.Hour),
		End:   currentTime.Add(time.Duration(4) * time.Hour),
	}

	f := NewFacility(FacilityName("Testing"))

	_ = f.Book(*b1)
	_ = f.Book(*b2)

	// Success case
	if err := f.UpdateBooking(b2.Id, 1); err != nil {
		t.Error("Error while updating booking")
	}
	if len(f.Bookings) != 2 {
		t.Error("Unexpected booking deletion upon update")
	}
	if f.Bookings[1].Start != b2.Start.Add(time.Duration(1)*time.Hour) || f.Bookings[1].End != b2.End.Add(time.Duration(1)*time.Hour) {
		t.Error("Booking timing was not updated properly")
	}

	// Error case
	if err := f.UpdateBooking(b2.Id, -2); err == nil {
		t.Error("Expected an error to be raised upon error update")
	}

}

func TestFacility_DeleteBooking(t *testing.T) {

	currentTime := time.Now()

	b1 := &Booking{
		Id:    1,
		Start: currentTime.Add(time.Duration(1) * time.Hour),
		End:   currentTime.Add(time.Duration(3) * time.Hour),
	}
	b2 := &Booking{
		Id:    2,
		Start: currentTime.Add(time.Duration(3) * time.Hour),
		End:   currentTime.Add(time.Duration(4) * time.Hour),
	}

	f := NewFacility(FacilityName("Testing"))

	_ = f.Book(*b1)
	_ = f.Book(*b2)

	if len(f.Bookings) != 2 {
		t.Error("Bookings were not added as expected")
	}

	f.DeleteBooking(b2.Id)

	if len(f.Bookings) != 1 {
		t.Error("Booking was not deleted")
	}

}
