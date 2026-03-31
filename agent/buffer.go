package agent

import (
	"iter"
	"slices"
	"sync"

	chkr "github.com/tweithoener/checker"
)

// EventBuffer stores a limited number of events using an efficient ring buffer.
// It caches the ordered slice to prevent multiple allocations on subsequent reads.
type EventBuffer struct {
	mu     sync.Mutex
	events []Event
	head   int
	tail   int
	count  int
	limit  int

	cachedEvents []Event
	isDirty      bool
}

// NewEventBuffer creates a new event buffer with the given capacity limit.
func NewEventBuffer(limit int) *EventBuffer {
	if limit <= 0 {
		limit = 100 // Safe default
	}
	return &EventBuffer{
		events:  make([]Event, limit),
		limit:   limit,
		isDirty: true,
	}
}

// Add appends a new event to the buffer. If full, the oldest event is overwritten.
func (b *EventBuffer) Add(cs chkr.CheckState) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ev := Event{
		Name:       cs.Name,
		CheckState: cs,
		ReceivedAt: TimeNow(),
	}

	b.events[b.tail] = ev
	b.tail = (b.tail + 1) % b.limit

	if b.count < b.limit {
		b.count++
	} else {
		// Overwrite the oldest element
		b.head = (b.head + 1) % b.limit
	}

	// Invalidate the cache whenever a new event is added
	b.isDirty = true
}

// Events returns a chronologically ordered iterator of the currently buffered events.
// It caches the underlying slice, making subsequent calls extremely fast and allocation-free
// until a new event is added. The returned iterator is safe for concurrent read access.
func (b *EventBuffer) Events() iter.Seq[Event] {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.isDirty {
		return slices.Values(b.cachedEvents)
	}

	res := make([]Event, 0, b.count)
	if b.count == 0 {
		b.cachedEvents = res
		b.isDirty = false
		return slices.Values(res)
	}

	if b.head < b.tail {
		res = append(res, b.events[b.head:b.tail]...)
	} else {
		res = append(res, b.events[b.head:b.limit]...)
		res = append(res, b.events[0:b.tail]...)
	}

	b.cachedEvents = res
	b.isDirty = false
	return slices.Values(res)
}

// Len returns the current number of events in the buffer.
func (b *EventBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.count
}
