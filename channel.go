package gosse

import (
	"io"
	"net/http"
	"sync"
)

type Channel struct {
	conns map[*SSEConn]struct{}
	mu    sync.RWMutex
}

func NewChannel() *Channel {
	return &Channel{
		conns: make(map[*SSEConn]struct{}),
	}
}

func (c *Channel) Subscribe(w io.Writer) *SSEConn {
	conn := NewSSEConn(w)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conns[conn] = struct{}{}
	return conn
}

func (c *Channel) Unsubscribe(conn *SSEConn) {
	c.mu.Lock()
	delete(c.conns, conn)
	c.mu.Unlock()
	conn.Close()
}

func (c *Channel) Push(e Event) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for conn := range c.conns {
		conn.Send(e)
	}
}

func (c *Channel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "event streaming not supported", http.StatusInternalServerError)
		return
	}

	conn := c.Subscribe(w)
	defer c.Unsubscribe(conn)

	flusher.Flush()

	for {
		select {
		case event, ok := <-conn.Listen():
			if !ok {
				return // closed channel
			}

			if _, err := conn.WriteEvent(event); err != nil {
				http.Error(w, "error trying to write an event", http.StatusInternalServerError)
				return
			}

			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (c *Channel) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for conn := range c.conns {
		conn.Close()
		delete(c.conns, conn)
	}
}
