package gosse

import (
	"errors"
	"net/http"
	"sync"
)

type Channel struct {
	config *ChannelConfig

	messages chan []byte
	closeCh  chan struct{}

	w  http.ResponseWriter
	rc *http.ResponseController

	mu  sync.Mutex
	err error
}

func NewChannel(w http.ResponseWriter, opts ...ChannelOpt) *Channel {
	config := DefaultChannelConfig()
	for _, opt := range opts {
		opt(config)
	}

	return &Channel{
		config:   config,
		messages: make(chan []byte, config.BufferSize),
		closeCh:  make(chan struct{}),
		w:        w,
		rc:       http.NewResponseController(w),
	}
}

// Send sends a new message to the channel.
func (c *Channel) Send(m []byte) (n int, err error) {
	c.mu.Lock()
	closed := c.closeCh == nil
	c.mu.Unlock()

	if closed {
		return 0, ErrChannelClosed
	}
	if c.messages != nil && m != nil {
		c.messages <- m
		return len(m), nil
	}
	return 0, nil
}

// Write writes a new payload to the underlying response writer.
func (c *Channel) Write(p []byte) (n int, err error) {
	if c.Closed() {
		return 0, ErrChannelClosed
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.w != nil {
		n, err := c.w.Write(p)
		if err != nil {
			return n, err
		}

		err = c.Flush()
		if err != nil {
			return n, err
		}

		return n, nil
	}
	return 0, nil
}

func (c *Channel) Close() error {
	if c.Closed() {
		return ErrChannelClosed
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closeCh != nil {
		close(c.closeCh)
		c.closeCh = nil
	}

	return c.err
}

// Listen blocks, invoking fn for every message received on the channel,
// until the Channel is closed.
func (c *Channel) Listen(fn func(msg []byte)) error {
	if c.Closed() {
		return ErrChannelClosed
	}
	for {
		select {
		case <-c.Done():
			return c.Error()
		case msg, ok := <-c.messages:
			if !ok {
				return c.Error()
			}
			fn(msg)
		}
	}
}

func (c *Channel) SetError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = errors.Join(c.err, err)
}

func (c *Channel) Error() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *Channel) Closed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closeCh == nil
}

// Done returns a channel that is closed once the Channel has been closed.
// If the Channel is already closed, it returns an already-closed channel,
// so callers doing `<-ch.Done()` never block forever.
func (c *Channel) Done() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closeCh == nil {
		done := make(chan struct{})
		close(done)
		return done
	}

	return c.closeCh
}

func (c *Channel) Flush() error {
	if c.rc != nil {
		return c.rc.Flush()
	}
	return nil
}
