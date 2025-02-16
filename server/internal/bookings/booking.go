package bookings

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"time"
)

type Booking struct {
	Id    uint16
	Start time.Time
	End   time.Time
}

type BookingOption func(*Booking)

func BookingWithRandomId() BookingOption {

	randomUint16 := uint16(rand.Uint32() & 0xFFFF) // Extract lower 16 bits

	return func(b *Booking) {
		b.Id = randomUint16
	}
}

func BookingWithStartTime(t time.Time) BookingOption {
	return func(b *Booking) {
		b.Start = t
	}
}

func BookingWithEndTime(t time.Time) BookingOption {
	return func(b *Booking) {
		b.End = t
	}
}

func NewBooking(opts ...BookingOption) (Booking, error) {
	b := &Booking{}
	for _, o := range opts {
		o(b)
	}

	if err := b.validate(); err != nil {
		return Booking{}, err
	}

	return *b, nil
}

func (b *Booking) validate() error {
	if b.Id == 0 || b.Start.IsZero() || b.End.IsZero() || b.End.Before(b.Start) || b.Start.Equal(b.End) {
		return errors.New("invalid configuration for Booking struct")
	}
	return nil
}

func (b *Booking) Overlaps(other *Booking) bool {
	return b.Start.Before(other.End) && other.Start.Before(b.End)
}

func (b *Booking) GetIdAsBytes() []byte {
	bin := make([]byte, 2)
	binary.BigEndian.PutUint16(bin, b.Id)
	return bin
}
