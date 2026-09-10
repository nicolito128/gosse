package gosse

import "errors"

var (
	ErrClientIsGone  = errors.New("client is gone")
	ErrSentFailed    = errors.New("sent failed")
	ErrFlushFailed   = errors.New("flush failed")
	ErrChannelClosed = errors.New("channel closed")
)
