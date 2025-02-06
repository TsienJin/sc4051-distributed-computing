package bookings

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// Manager is responsible for handling facilities, bookings and monitoring.
type Manager struct {
	sync.RWMutex
	Facilities map[FacilityName]*Facility
	monitor    *Monitor
}

var (
	manager     *Manager
	onceManager sync.Once
)

func GetManager() *Manager {

	onceManager.Do(func() {
		manager = &Manager{
			Facilities: make(map[FacilityName]*Facility),
			monitor:    GetMonitor(),
		}
	})

	return manager
}

func (m *Manager) NewFacility(name FacilityName) error {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.Facilities[name]; exists {
		slog.Error("Attempted to create a Facility that already exists!", "Facility", name)
		return errors.New("facility already exists")
	}
	m.Facilities[name] = NewFacility(name)
	return nil
}

func (m *Manager) QueryFacility(n FacilityName, days int) ([]byte, error) {
	if _, exists := m.Facilities[n]; !exists {
		slog.Error("Facility does not exist!", "FacilityName", n)
		return []byte{}, errors.New("facility does not exist")
	}
	m.monitor.Update(n, fmt.Sprintf("Executing query on %s for %d days", n, days))
	return m.Facilities[n].QueryAvailability(days), nil
}

func (m *Manager) DeleteFacility(name FacilityName) error {
	m.Lock()
	defer m.Unlock()

	// Check if it exists
	if r, exists := m.Facilities[name]; !exists || len(r.Bookings) > 0 {
		switch {
		case !exists:
			slog.Error("Attempted to delete a Facility that does not exists!", "Facility", name)
			return errors.New("facility does not exists")
		case len(r.Bookings) > 0:
			slog.Error("Attempted to delete a Facility with existing bookings!", "Facility", name, "bookings", r)
			m.monitor.Update(name, "A deletion was attempted on this facility! There are existing bookings can cannot be deleted yet.")
			return errors.New("facility has existing bookings")
		}
	}

	// "OK" to delete at this point
	delete(m.Facilities, name)
	m.monitor.Update(name, "This facility has been deleted.")
	slog.Info("Deleted facility", "FacilityName", name)
	m.monitor.Clear(name)
	return nil
}

func (m *Manager) NewBooking(n FacilityName, b Booking) error {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.Facilities[n]; !exists {
		slog.Error("Attempted to book a Facility that does not exists!", "FacilityName", n)
		return errors.New("facility does not exists")
	}
	if err := m.Facilities[n].Book(b); err != nil {
		slog.Error("Unable to make booking", "FacilityName", n, "Booking", b)
		m.monitor.Update(n, fmt.Sprintf("Error attempting to make booking at %s with %v.", n, b))
		return err
	}
	slog.Info("Made successful booking", "FacilityName", n, "Booking", b)
	m.monitor.Update(n, fmt.Sprintf("Successfully made booking at %s with %v", n, b))
	return nil
}

func (m *Manager) UpdateBooking(n FacilityName, bookingId uint16, deltaHours int) error {
	m.RLock()
	defer m.RUnlock()

	if _, exists := m.Facilities[n]; !exists {
		slog.Error("Facility does not exist!", "FacilityName", n)
		return errors.New("facility does not exist")
	}
	if err := m.Facilities[n].UpdateBooking(bookingId, deltaHours); err != nil {
		slog.Error("Failed to update booking!", "BookingId", bookingId, "DeltaHours", deltaHours)
		m.monitor.Update(n, fmt.Sprintf("Failed to update BookingId %v by %d hours.", bookingId, deltaHours))
		return err
	}
	m.monitor.Update(n, fmt.Sprintf("Updated BookingId %v by %d hours", bookingId, deltaHours))
	return nil
}

func (m *Manager) UpdateBookingFromId(id uint16, deltaHours int) error {
	m.RLock()
	defer m.RUnlock()

	updated := false

	for _, f := range m.Facilities {
		if f.HasId(id) {
			if err := m.UpdateBooking(f.Name, id, deltaHours); err != nil {
				return err
			}
			updated = true
		}
	}

	if !updated {
		slog.Error("Booking with Id not found!", "BookingId", id)
		return errors.New("booking with Id not found")
	}

	return nil
}

func (m *Manager) DeleteBooking(n FacilityName, bookingId uint16) error {
	m.RLock()
	defer m.RUnlock()

	if _, exists := m.Facilities[n]; !exists {
		slog.Error("Facility does not exist!", "FacilityName", n)
		return errors.New("facility does not exist")
	}

	if deleted := m.Facilities[n].DeleteBooking(bookingId); deleted {
		slog.Info("Deleted booking", "BookingId", bookingId)
		m.monitor.Update(n, fmt.Sprintf("Successfully deleted Booking %X from %s.", bookingId, n))
	} else {
		slog.Warn("Attempted to delete non-existent booking", "BookingId", bookingId)
		m.monitor.Update(n, fmt.Sprintf("Attempted to delete non-existant Booking %X from %s.", bookingId, n))
	}

	return nil
}

func (m *Manager) DeleteBookingFromId(id uint16) error {
	m.RLock()
	defer m.RUnlock()

	deleted := false

	for _, f := range m.Facilities {
		if f.HasId(id) {
			if err := m.DeleteBooking(f.Name, id); err != nil {
				return err
			}
			deleted = true
		}
	}

	if !deleted {
		slog.Error("Booking with Id not found!", "BookingId", id)
		return errors.New("booking with Id not found")
	}

	return nil
}
