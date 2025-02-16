package bookings

import (
	"fmt"
	"testing"
	"time"
)

func TestManager_NewBooking(t *testing.T) {
	manager := GetManager()
	facilityName := FacilityName("TestManager_NewBooking")

	currentTime := time.Now()

	if err := manager.NewFacility(facilityName); err != nil {
		t.Error(err)
	}

	b1, err := NewBooking(
		BookingWithRandomId(),
		BookingWithStartTime(currentTime.Add(time.Duration(1)*time.Hour)),
		BookingWithEndTime(currentTime.Add(time.Duration(3)*time.Hour)),
	)

	if err != nil {
		t.Error(err)
	}

	if err := manager.NewBooking(facilityName, b1); err != nil {
		t.Error(err)
	}
}

func TestManager_NewBooking_fail_duplicate(t *testing.T) {
	manager := GetManager()
	facilityName := FacilityName("TestManager_NewBooking_fail_duplicate")

	currentTime := time.Now()

	if err := manager.NewFacility(facilityName); err != nil {
		t.Error(err)
	}

	b1, err := NewBooking(
		BookingWithRandomId(),
		BookingWithStartTime(currentTime.Add(time.Duration(1)*time.Hour)),
		BookingWithEndTime(currentTime.Add(time.Duration(3)*time.Hour)),
	)

	if err != nil {
		t.Error(err)
	}

	if err := manager.NewBooking(facilityName, b1); err != nil {
		t.Error(err)
	}

	if err := manager.NewBooking(facilityName, b1); err == nil {
		t.Error(fmt.Errorf("expected booking to fail since it is a duplicate"))
	}
}

func TestManager_NewBooking_fail_clashing(t *testing.T) {
	manager := GetManager()
	facilityName := FacilityName("TestManager_NewBooking_fail_clashing")

	currentTime := time.Now()

	if err := manager.NewFacility(facilityName); err != nil {
		t.Error(err)
	}

	b1, err := NewBooking(
		BookingWithRandomId(),
		BookingWithStartTime(currentTime.Add(time.Duration(1)*time.Hour)),
		BookingWithEndTime(currentTime.Add(time.Duration(3)*time.Hour)),
	)

	if err != nil {
		t.Error(err)
	}

	b2, err := NewBooking(
		BookingWithRandomId(),
		BookingWithStartTime(currentTime.Add(time.Duration(1)*time.Hour)),
		BookingWithEndTime(currentTime.Add(time.Duration(3)*time.Hour)),
	)

	if err != nil {
		t.Error(err)
	}

	if err := manager.NewBooking(facilityName, b1); err != nil {
		t.Error(err)
	}

	if err := manager.NewBooking(facilityName, b2); err == nil {
		t.Error(fmt.Errorf("expected booking to fail since it is clashing"))
	}
}

func TestManager_DeleteBookingFromId(t *testing.T) {

	manager := GetManager()
	facilityName := FacilityName("TestManager_DeleteBookingFromId")

	currentTime := time.Now()

	if err := manager.NewFacility(facilityName); err != nil {
		t.Error(err)
	}

	b1 := &Booking{
		Id:    1,
		Start: currentTime.Add(time.Duration(1) * time.Hour),
		End:   currentTime.Add(time.Duration(3) * time.Hour),
	}

	if err := manager.NewBooking(facilityName, *b1); err != nil {
		t.Error(err)
	}

	if err := manager.DeleteBookingFromId(b1.Id); err != nil {
		t.Error(err)
	}
}
