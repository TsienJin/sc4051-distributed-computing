package bookings

import (
	"errors"
	"log/slog"
	"slices"
	"sync"
	"time"
)

type FacilityName string

type Facility struct {
	sync.RWMutex
	Name       FacilityName
	Bookings   []*Booking
	BookingMap map[uint16]*Booking
}

func NewFacility(name FacilityName) *Facility {
	return &Facility{
		Name:       name,
		Bookings:   []*Booking{},
		BookingMap: make(map[uint16]*Booking),
	}
}

// clean clears up outdated bookings from the system
func (f *Facility) clean() {

	currentTime := time.Now()

	// Remove bookings that have already finished
	f.Bookings = slices.DeleteFunc(f.Bookings, func(b *Booking) bool {
		return b.End.Before(currentTime)
	})
}

// insertBooking attempts to insert a booking in sorted order
func (f *Facility) insertBooking(newBooking *Booking) bool {
	// Find the correct insertion index
	index, found := slices.BinarySearchFunc(f.Bookings, newBooking, func(a, b *Booking) int {
		if a.Start.Before(b.Start) {
			return -1
		}
		if a.Start.After(b.Start) {
			return 1
		}
		return 0
	})

	// Ensure no overlap
	if found || (index > 0 && f.Bookings[index-1].Overlaps(newBooking)) ||
		(index < len(f.Bookings) && f.Bookings[index].Overlaps(newBooking)) {
		return false // Conflict detected
	}

	// Insert at the correct position
	f.Bookings = append(f.Bookings[:index], append([]*Booking{newBooking}, f.Bookings[index:]...)...)
	f.BookingMap[newBooking.Id] = newBooking
	return true
}

// QueryAvailability searches for availability of the facility for the next number of nDays (including today)
// Returns:
// - []byte, where each bit represents the availability of the facility corresponding to the hour.
func (f *Facility) QueryAvailability(nDays int) []byte {
	f.Lock()
	defer f.Unlock()
	f.clean()

	schedule := make([]byte, nDays*3)

	currentTime := time.Now()
	firstDate := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.Local)

	for _, b := range f.Bookings {
		if b.End.Before(firstDate) {
			continue
		}

		firstHour := int(b.Start.Sub(firstDate).Hours())
		lastHour := int(b.End.Sub(firstDate).Hours())

		if firstHour < 0 {
			firstHour = 0 // floor
		}
		if lastHour >= nDays*24 {
			lastHour = nDays*24 - 1 // Booking ends after the queried period
		}

		// Set bits for each hour in the schedule
		for hour := firstHour; hour < lastHour; hour++ {
			byteIndex := hour / 8                // of the 3 bytes per, which byte is responsible for this timing
			bitIndex := 7 - (hour % 8)           // of the 8 bits, which bit is responsible for the hour
			schedule[byteIndex] |= 1 << bitIndex // set corresponding hour to be 1
		}

	}

	return schedule
}

func (f *Facility) HasId(id uint16) bool {
	f.RLock()
	defer f.RUnlock()

	_, exists := f.BookingMap[id]
	return exists
}

func (f *Facility) Book(b Booking) error {
	f.Lock()
	defer f.Unlock()
	f.clean()

	if !f.insertBooking(&b) {
		slog.Error("Unable to insert booking due to clashes")
		return errors.New("unable to insert booking due to clashes")
	}

	return nil
}

func (f *Facility) UpdateBooking(id uint16, deltaHours int) error {
	f.Lock()
	defer f.Unlock()
	f.clean()

	// Get index of ID
	index := -1
	var booking Booking
	for i, b := range f.Bookings {
		if b.Id == id {
			index = i
			booking = *b
			break
		}
	}
	if index == -1 {
		slog.Error("Unable to find booking to be updated")
		return errors.New("booking with specified ID not found")
	}

	// Remove existing booking
	f.Bookings = append(f.Bookings[:index], f.Bookings[index+1:]...)

	// Updating booking timing
	newBooking := booking // create a copy of the existing booking
	newBooking.Start = newBooking.Start.Add(time.Duration(deltaHours) * time.Hour)
	newBooking.End = newBooking.End.Add(time.Duration(deltaHours) * time.Hour)

	// Attempt to insert the updated booking
	if ok := f.insertBooking(&newBooking); !ok {
		// Fall back to original booking
		_ = f.insertBooking(&booking)
		slog.Error("Unable to update booking")
		return errors.New("unable to modify booking due to clashes")
	}

	slog.Info("Booking has been updated", "OriginalBooking", booking, "NewBooking", newBooking)
	return nil
}

func (f *Facility) DeleteBooking(id uint16) bool {
	f.Lock()
	defer f.Unlock()
	f.clean()

	deleted := false

	// Remove bookings that match Id
	f.Bookings = slices.DeleteFunc(f.Bookings, func(b *Booking) bool {
		if b.Id == id {
			delete(f.BookingMap, b.Id)
			deleted = true
			return true
		}
		return false
	})

	return deleted
}
