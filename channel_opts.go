package gosse

const (
	DefaultBufferSize = 1024
)

type ChannelOpt func(*ChannelConfig)

type ChannelConfig struct {
	BufferSize int
}

func DefaultChannelConfig() *ChannelConfig {
	return &ChannelConfig{
		BufferSize: DefaultBufferSize,
	}
}

func WithBufferSize(size int) ChannelOpt {
	return func(c *ChannelConfig) {
		if size > 0 {
			c.BufferSize = size
		}
	}
}
