package client

import (
	"fmt"
	"log/slog"
	"server/internal/protocol/proto_defs"
	"server/internal/rpc/response"
	"sync"
)

type responseSequencer struct {
	sync.RWMutex

	cached    map[proto_defs.MessageId]*response.Response
	sequence  []proto_defs.MessageId
	completed map[proto_defs.MessageId]struct{}

	targetChan chan *response.Response
}

func newResponseSequencer(c chan *response.Response) *responseSequencer {
	return &responseSequencer{
		cached:     make(map[proto_defs.MessageId]*response.Response),
		sequence:   []proto_defs.MessageId{},
		completed:  make(map[proto_defs.MessageId]struct{}),
		targetChan: c,
	}
}

func (r *responseSequencer) expect(id proto_defs.MessageId) {
	r.Lock()
	defer r.Unlock()

	slog.Info("[SEQUENCER] Setting order for new packet")
	r.sequence = append(r.sequence, id)
}

func (r *responseSequencer) handle(id proto_defs.MessageId, res *response.Response) {
	r.Lock()
	defer r.Unlock()

	slog.Info("[SEQUENCER] Handling response")

	if _, exists := r.completed[id]; exists {
		slog.Info("[SEQUENCER] Packet has been handled, exiting")
		return
	}

	r.cached[id] = res

	var i int
	for i = 0; i < len(r.sequence); i++ {
		cRes, exists := r.cached[r.sequence[i]]
		if !exists {
			slog.Info(fmt.Sprintf("[SEQUENCER] Current packets in expected: %v", r.sequence))
			slog.Info(fmt.Sprintf("[SEQUENCER] Packet %v of %v total expected is missing! Exiting", i, len(r.sequence)))
			slog.Info(fmt.Sprintf("[SEQUENCER] Current Id: %v", r.sequence[i]))
			slog.Info(fmt.Sprintf("[SEQUENCER] cRes for Id: %v", cRes))
			break
		}

		r.completed[r.sequence[i]] = struct{}{}
		slog.Info("[SEQUENCER] Packet has been handed off")
		r.targetChan <- cRes
		delete(r.cached, r.sequence[i])

	}

	r.sequence = r.sequence[i:]

}
