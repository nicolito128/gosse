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
	defer c.mu.Unlock()
	delete(c.conns, conn)
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

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "event streaming not supported", http.StatusInternalServerError)
		return
	}

	conn := c.Subscribe(w)
	defer c.Unsubscribe(conn)

	for {
		select {
		case event := <-conn.Listen():
			_, err := conn.WriteEvent(event)
			if err != nil {
				http.Error(w, "error trying to write an event", http.StatusInternalServerError)
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
