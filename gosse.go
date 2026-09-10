package gosse

import (
	"net/http"
)

func Upgrade(w http.ResponseWriter, r *http.Request, opts ...ChannelOpt) *Channel {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	gone := r.Context().Done()
	c := NewChannel(w, opts...)

	go func() {
		defer c.Close()

		for {
			select {
			case <-gone:
				c.SetError(ErrClientIsGone)
				return

			case <-c.Done():
				return

			case msg := <-c.messages:
				_, err := c.Write(msg)
				if err != nil {
					c.SetError(ErrSentFailed)
				}
			}
		}
	}()

	return c
}
