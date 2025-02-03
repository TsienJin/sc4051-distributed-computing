package bookings

import (
	"testing"
	"time"
)

func TestBooking_Overlaps(t *testing.T) {
	b1 := &Booking{
		Id:    1,
		Start: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local),
		End:   time.Date(2025, time.January, 1, 3, 0, 0, 0, time.Local),
	}
	b2 := &Booking{
		Id:    2,
		Start: time.Date(2025, time.January, 1, 3, 0, 0, 0, time.Local),
		End:   time.Date(2025, time.January, 1, 4, 0, 0, 0, time.Local),
	}
	b3 := &Booking{
		Id:    3,
		Start: time.Date(2025, time.January, 1, 1, 0, 0, 0, time.Local),
		End:   time.Date(2025, time.January, 1, 4, 0, 0, 0, time.Local),
	}

	if b1.Overlaps(b2) {
		t.Error("Bookings B1 and B2 should not overlap")
	}

	if !b3.Overlaps(b1) || !b3.Overlaps(b2) {
		t.Error("Bookings B3 should overlap with B1 or B2")
	}
}
