package monitor

import (
	"log/slog"
	"sync"
)

type networkStats struct {
	mu                sync.RWMutex
	packetInExpected  int // Total number of packets that the server actual receives
	packetInDropped   int // Number of inbound packets that have been dropped
	packetOutExpected int // Total number of packets that the server supposed to send out
	packetOutDropped  int // Number of outbound packets that have been dropped
}

var (
	n                *networkStats
	networkStatsOnce sync.Once
)

func init() {
	networkStatsOnce.Do(func() {
		n = &networkStats{
			mu:                sync.RWMutex{},
			packetInExpected:  0,
			packetInDropped:   0,
			packetOutExpected: 0,
			packetOutDropped:  0,
		}
	})
}

func MarkPacketIn() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.packetInExpected++
}

func MarkPacketInDropped() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.packetInDropped++
}

func MarkPacketOut() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.packetOutExpected++
}

func MarkPacketOutDropped() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.packetOutDropped++
}

func resetNetworkMonitor() {
	n.mu.Lock()
	defer n.mu.Unlock()
	slog.Warn("Resetting network monitor")
	n.packetInExpected = 0
	n.packetInDropped = 0
	n.packetOutExpected = 0
	n.packetOutDropped = 0
}

func getNetworkStats() networkStats {
	return networkStats{
		mu:                sync.RWMutex{},
		packetInExpected:  n.packetInExpected,
		packetInDropped:   n.packetInDropped,
		packetOutExpected: n.packetOutExpected,
		packetOutDropped:  n.packetOutDropped,
	}
}
