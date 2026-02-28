package gosse

import (
	"fmt"
	"io"
)

type SSEConn struct {
	io.Writer
	ch     chan Event
	closed bool
}

func NewSSEConn(w io.Writer) *SSEConn {
	ch := make(chan Event, 16)
	return &SSEConn{
		Writer: w,
		ch:     ch,
	}
}

func (sc *SSEConn) ok() bool {
	return sc.ch != nil && !sc.closed
}

func (sc *SSEConn) WriteEvent(e Event) (n int, err error) {
	if !sc.ok() {
		return 0, fmt.Errorf("sse connection not started")
	}
	n, err = sc.Write([]byte(e.String()))
	return
}

func (sc *SSEConn) Send(e Event) {
	select {
	case sc.ch <- e:
	default:
	}
}

func (sc *SSEConn) Listen() <-chan Event {
	return sc.ch
}

func (sc *SSEConn) Close() error {
	if !sc.ok() {
		return fmt.Errorf("sse connection already closed")
	}
	close(sc.ch)
	sc.closed = true
	return nil
}
