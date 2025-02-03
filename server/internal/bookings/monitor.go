package bookings

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"
)

type MonitorConsumer struct {
	Channel chan string
	Cancel  context.CancelFunc
	closed  sync.Once
}

type Monitor struct {
	sync.Mutex
	Watchers map[FacilityName][]*MonitorConsumer
}

var (
	monitor     *Monitor
	onceMonitor sync.Once
)

func NewMonitor() *Monitor {
	onceMonitor.Do(func() {
		monitor = &Monitor{
			Watchers: make(map[FacilityName][]*MonitorConsumer),
		}
	})
	return monitor
}

func (m *Monitor) Watch(f FacilityName, ttl time.Duration) *MonitorConsumer {

	updateChannel := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), ttl)

	consumer := &MonitorConsumer{
		Channel: updateChannel,
		Cancel:  cancel,
	}

	m.Lock()
	m.Watchers[f] = append(m.Watchers[f], consumer)
	m.Unlock()

	go func() {
		slog.Info("Created MonitorConsumer", "facility", f)

		<-ctx.Done()
		slog.Info("MonitorConsumer expired or terminated, removing from watchers")

		consumer.closed.Do(func() {
			close(consumer.Channel)
		})

		// Remove watcher
		m.Lock()
		m.Watchers[f] = slices.DeleteFunc(m.Watchers[f], func(monitorConsumer *MonitorConsumer) bool {
			return monitorConsumer == consumer
		})
		m.Unlock()
	}()

	return consumer
}

func (m *Monitor) Update(f FacilityName, message string) {
	m.Lock()
	defer m.Unlock()

	// Silently exit if no known watches for facility
	if _, e := m.Watchers[f]; !e {
		return
	}

	slog.Info("Sending messages to watchers", "FacilityName", f, "message", message)

	// Send message to all watchers
	for _, w := range m.Watchers[f] {
		w.Channel <- message
	}

}

func (m *Monitor) Clear(f FacilityName) {
	m.Lock()
	defer m.Unlock()

	// Silently exit if no known watches for facility
	if _, e := m.Watchers[f]; !e {
		return
	}

	slog.Info("Clearing watches of facility", "FacilityName", f)

	// Transmit one more message before closing channel
	for _, w := range m.Watchers[f] {
		w.closed.Do(func() {
			w.Channel <- "Terminating transmission."
			close(w.Channel)
			w.Cancel()
		})
	}

}
