package events

import "sync"

type EventType string

const (
	EventForegroundChanged EventType = "foreground_changed"
	EventWindowMoved       EventType = "window_moved"
	EventWindowShown       EventType = "window_shown"
	EventWindowHidden      EventType = "window_hidden"
)

type SystemEvent struct {
	Type      EventType `json:"type"`
	Timestamp int64     `json:"timestampUTC"`
	PID       int       `json:"pid"`
	HWND      uintptr   `json:"hwnd"`
}

type Bus struct {
	ch   chan SystemEvent
	once sync.Once
	mu   sync.Mutex
	win  *windowsSources
}

func NewBus(buffer int) *Bus {
	return &Bus{ch: make(chan SystemEvent, buffer)}
}

func (b *Bus) Events() <-chan SystemEvent { return b.ch }

// Emit never blocks: under a flood the oldest pending events are the ones that
// matter least, and the capture loop debounces per process anyway.
func (b *Bus) Emit(ev SystemEvent) {
	select {
	case b.ch <- ev:
	default:
	}
}

func (b *Bus) StartWindowsSources() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.win != nil {
		return
	}
	b.win = newWindowsSources(b.Emit)
	b.win.start()
}

func (b *Bus) Stop() {
	b.once.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if b.win != nil {
			b.win.stop()
		}
	})
}
