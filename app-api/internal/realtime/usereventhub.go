package realtime

import (
	"sync"

	"wklive/common/userevent"
)

type UserEvent = userevent.Event

type userKey struct {
	tenantID int64
	userID   int64
}

type UserEventHub struct {
	mu          sync.RWMutex
	subscribers map[userKey]map[chan UserEvent]struct{}
}

func NewUserEventHub() *UserEventHub {
	return &UserEventHub{subscribers: make(map[userKey]map[chan UserEvent]struct{})}
}

func (h *UserEventHub) Subscribe(tenantID, userID int64) (<-chan UserEvent, func()) {
	key := userKey{tenantID: tenantID, userID: userID}
	channel := make(chan UserEvent, 32)

	h.mu.Lock()
	if h.subscribers[key] == nil {
		h.subscribers[key] = make(map[chan UserEvent]struct{})
	}
	h.subscribers[key][channel] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subscribers[key], channel)
			if len(h.subscribers[key]) == 0 {
				delete(h.subscribers, key)
			}
			close(channel)
			h.mu.Unlock()
		})
	}
	return channel, cancel
}

func (h *UserEventHub) Publish(event UserEvent) {
	if event.TenantID <= 0 || event.UserID <= 0 {
		return
	}
	key := userKey{tenantID: event.TenantID, userID: event.UserID}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for channel := range h.subscribers[key] {
		select {
		case channel <- event:
		default:
			// A slow connection receives a resync notification on reconnect and
			// HTTP remains the final source of truth.
		}
	}
}
