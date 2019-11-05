package server

import "github.com/sereiner/duo/codec"

type Option func(*option)

type option struct {
	codecType codec.CodecType
}
