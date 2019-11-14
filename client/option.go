package client

import (
	"github.com/sereiner/duo/codec"
	"time"
)

type Option func(*option)

type option struct {
	codecType      codec.CodecType
	RequestTimeout time.Duration
}

func WithCodecType(c codec.CodecType) Option {

	return func(o *option) {
		o.codecType = c
	}
}

func WithRequestTimeout(t time.Duration) Option {
	return func(o *option) {
		o.RequestTimeout = t
	}
}
