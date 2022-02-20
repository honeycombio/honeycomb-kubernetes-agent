package transmission

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
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

	if r.size > 0 && r.ttl > 0 {
		r.startCleanupTimer()
	}

	return r
}

// Add an event to ring buffer
func (r *RingBuffer) Add(key uint64, ev *event.Event) {
	// fast return if disabled
	if r.size == 0 {
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
		if _, ok := r.items[oldKey]; ok {
			// event exists, remove it from the items map
			delete(r.items, oldKey)
		}
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
	if r.size == 0 {
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
	ticker := time.Tick(duration)
	go (func() {
		for {
			select {
			case <-ticker:
				r.cleanup()
			}
		}
	})()
}
