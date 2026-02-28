package gosse

import (
	"fmt"
	"io"
	"net"
	"time"
)

type HandlerFunc func(*SSEConn)

type SSEConn struct {
	io.Writer
	ch     chan Event
	closed bool
	conn   net.Conn
}

func NewSSEConn(w io.Writer) *SSEConn {
	ch := make(chan Event, 16)
	sc := &SSEConn{
		Writer: w,
		ch:     ch,
	}

	if c, ok := w.(net.Conn); ok {
		sc.conn = c
	}

	return sc
}

func (sc *SSEConn) ok() bool {
	return sc.ch != nil && !sc.closed
}

func (sc *SSEConn) Read(b []byte) (n int, err error) {
	if sc.conn == nil {
		return 0, fmt.Errorf("read not supported: underlying writer is not a net.Conn")
	}
	return sc.conn.Read(b)
}

func (sc *SSEConn) LocalAddr() net.Addr {
	if sc.conn == nil {
		return nil
	}
	return sc.conn.LocalAddr()
}

func (sc *SSEConn) RemoteAddr() net.Addr {
	if sc.conn == nil {
		return nil
	}
	return sc.conn.RemoteAddr()
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

func (sc *SSEConn) SetDeadline(t time.Time) error {
	if sc.conn == nil {
		return fmt.Errorf("deadlines not supported")
	}
	return sc.conn.SetDeadline(t)
}

func (sc *SSEConn) SetReadDeadline(t time.Time) error {
	if sc.conn == nil {
		return fmt.Errorf("read deadlines not supported")
	}
	return sc.conn.SetReadDeadline(t)
}

func (sc *SSEConn) SetWriteDeadline(t time.Time) error {
	if sc.conn == nil {
		return fmt.Errorf("underlying writer does not support deadlines")
	}
	return sc.conn.SetWriteDeadline(t)
}

func (sc *SSEConn) Close() error {
	if !sc.ok() {
		return fmt.Errorf("sse connection already closed")
	}

	var err error
	if sc.conn != nil {
		err = sc.conn.Close()
	}

	close(sc.ch)
	sc.closed = true
	return err
}
