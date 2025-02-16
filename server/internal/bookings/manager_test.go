package bookings

import (
	"testing"
	"time"
)

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
