package transmission

import (
	"sync"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/sirupsen/logrus"
)

type BufferEvent struct {
	*event.Event
	expires *time.Time
}

func (be *BufferEvent) expired() bool {
	if be.expires != nil {
		return be.expires.Before(time.Now())
	}
	return true
}

type RingBuffer struct {
	sync.RWMutex

	items     map[uint64]*BufferEvent
	itemOrder []uint64
	size      int
	position  int
	ttl       time.Duration
}

func NewRingBuffer(size int, duration time.Duration) *RingBuffer {
	logrus.WithFields(logrus.Fields{
		"size":    size,
		"timeout": duration,
	}).Info("Creating retry buffer.")

	r := &RingBuffer{
		items:     make(map[uint64]*BufferEvent),
		itemOrder: make([]uint64, size),
		size:      size,
		ttl:       duration,
	}

	if r.enabled() && r.ttl > 0 {
		r.startCleanupTimer()
	}

	return r
}

func (r *RingBuffer) enabled() bool {
	return r.size > 0
}

// Add an event to ring buffer
func (r *RingBuffer) Add(key uint64, ev *event.Event) {
	// fast return if disabled
	if !r.enabled() {
		return
	}

	expiration := time.Now().Add(r.ttl)
	be := &BufferEvent{
		Event:   ev,
		expires: &expiration,
	}

	r.Lock()
	// check if we have an existing event at this position
	oldKey := r.itemOrder[r.position]
	if oldKey != 0 {
		// If the key at the current position is associated with an event
		// in the buffer, remove it. Otherwise, this delete is a NOOP.
		delete(r.items, oldKey)
	}

	// add event to map and order array
	r.items[key] = be
	r.itemOrder[r.position] = key

	// increment position
	r.position++
	if r.position >= r.size {
		r.position = 0
	}
	r.Unlock()
}

// Get an event from the buffer's map
func (r *RingBuffer) Get(key uint64) (*event.Event, bool) {
	// fast return if disabled
	if !r.enabled() {
		return nil, false
	}

	r.RLock()
	be, ok := r.items[key]
	r.RUnlock()
	if ok {
		return be.Event, true
	} else {
		return nil, false
	}
}

func (r *RingBuffer) cleanup() {
	r.Lock()
	beforeSize := len(r.items)
	for key, be := range r.items {
		if be.expired() {
			delete(r.items, key)
		}
	}
	afterSize := len(r.items)
	r.Unlock()

	logrus.WithFields(logrus.Fields{
		"beforeSize": beforeSize,
		"afterSize":  afterSize,
	}).Debug("Retry buffer cleanup")
}

func (r *RingBuffer) startCleanupTimer() {
	duration := r.ttl
	if duration < time.Second {
		duration = time.Second
	}
	go (func() {
		for range time.Tick(duration) {
			r.cleanup()
		}
	})()
}
