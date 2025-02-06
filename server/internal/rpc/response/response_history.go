package response

import (
	"errors"
	"fmt"
	"log/slog"
	"server/internal/protocol/proto_defs"
	"server/internal/vars"
	"sync"
	"time"
)

type HistoryRecord struct {
	sync.RWMutex
	Response *Response
	Updated  time.Time
}

func NewHistoryRecord(r *Response) *HistoryRecord {
	return &HistoryRecord{
		Response: r,
		Updated:  time.Now(),
	}
}

func (h *HistoryRecord) GetResponse() *Response {
	h.Lock()
	defer h.Unlock()
	h.Updated = time.Now()
	return h.Response
}

type History struct {
	sync.RWMutex
	responses map[proto_defs.MessageId]*HistoryRecord
}

var (
	history     *History
	onceHistory sync.Once
)

func GetResponseHistoryInstance() *History {
	onceHistory.Do(func() {
		history = &History{responses: make(map[proto_defs.MessageId]*HistoryRecord)}

		t := time.NewTicker(time.Duration(vars.GetStaticEnv().ResponseIntervals) * time.Millisecond)
		go func() {
			defer t.Stop()
			for range t.C {
				history.CleanUp()
			}
		}()

	})

	return history
}

func (h *History) CleanUp() {
	h.Lock()
	defer h.Unlock()

	slog.Info("Cleaning up expired responses")
	count := 0

	expiredTime := time.Now().Add(-time.Duration(vars.GetStaticEnv().ResponseTTL) * time.Millisecond)

	for id, r := range h.responses {
		if r.Updated.Before(expiredTime) {
			count++
			delete(h.responses, id)
		}
	}
	slog.Info(fmt.Sprintf("Cleaned up %d expired responses", count))
}

func (h *History) SetProcessing(m proto_defs.MessageId) {
	h.Lock()
	defer h.Unlock()

	h.responses[m] = NewHistoryRecord(nil)
}

func (h *History) AddResponse(r *Response) {
	h.Lock()
	defer h.Unlock()

	h.responses[r.OriginalMessageId] = NewHistoryRecord(r)
}

// Check returns (response has been recorded, recorded into system)
// Whereby:
// - [0] == true => response has already been generated
// - [1] == true => request has already been received
func (h *History) Check(id proto_defs.MessageId) (bool, bool) {
	h.RLock()
	defer h.RUnlock()
	r, exists := h.responses[id]
	return r != nil, exists
}

func (h *History) GetResponse(id proto_defs.MessageId) (*Response, error) {
	h.Lock()
	defer h.Unlock()

	if r, exists := h.responses[id]; exists {

		if r.Response == nil {
			slog.Warn("Requested for response that has yet to be set", "RequestedMessageId", id)
			return nil, errors.New("response not marked as complete")
		}

		return r.GetResponse(), nil
	}

	slog.Warn("Requested for response that does not exist", "RequestedMessageId", id)
	return nil, errors.New("response does not exists")
}

func (h *History) RemoveResponse(id proto_defs.MessageId) {
	h.Lock()
	defer h.Unlock()
	delete(h.responses, id)
}
