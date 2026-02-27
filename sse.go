/*
Package gosse ...
*/
package gosse

import (
	"fmt"
	"net/http"
	"sync"
)

// Event ...
// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#fields
type Event struct {
	ID    string `json:"id"`
	Event string `json:"event"`
	Data  string `json:"data"`
	Retry int    `json:"retry"`
}

func (e Event) String() string {
	var s string
	if e.ID != "" {
		s += fmt.Sprintf("id: %s\n", e.ID)
	}

	if e.Event != "" {
		s += fmt.Sprintf("event: %s\n", e.Event)
	}

	if e.Retry > 0 {
		s += fmt.Sprintf("retry: %d\n", e.Retry)
	}

	s += fmt.Sprintf("data: %s\n\n", e.Data)

	return s
}

type Channel struct {
	conns map[chan Event]struct{}
	mu    sync.RWMutex
}

func NewChannel() *Channel {
	return &Channel{
		conns: make(map[chan Event]struct{}),
	}
}

func (c *Channel) Subscribe() chan Event {
	ch := make(chan Event, 16)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conns[ch] = struct{}{}
	return ch
}

func (c *Channel) Unsubscribe(ch chan Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.conns, ch)
	close(ch)
}

func (c *Channel) Push(event Event) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for ch := range c.conns {
		select {
		case ch <- event:
		default:
		}
	}
}

func (c *Channel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "event streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := c.Subscribe()
	defer c.Unsubscribe(ch)

	for {
		select {
		case event := <-ch:
			fmt.Fprint(w, event.String())
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
