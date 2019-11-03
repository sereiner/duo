package codec

import (
	"fmt"

	"github.com/sereiner/library/concurrent/cmap"
)

type CodecType byte

const (
	MsgPackCodecType CodecType = iota + 1
	GobCodecType
)

func (c CodecType) String() string {
	switch c {
	case MsgPackCodecType:

		return "序列化组件 messagepack"

	case GobCodecType:

		return "序列化组件 golang/gob"

	default:

		return "未知的二进制序列化组件 unknown"
	}
}

type Codec interface {
	Encode(value interface{}) ([]byte, error)
	Decode(data []byte, value interface{}) error
}

var codecMap cmap.ConcurrentMap

func init() {
	codecMap = cmap.New(2)
}

type Resolver interface {
	Resolve(name CodecType) (Codec, error)
}

var codecResolvers = make(map[CodecType]Resolver)

func Register(name CodecType, resolver Resolver) {
	if resolver == nil {
		panic("codec: resolver adapter is nil")
	}
	if _, ok := codecResolvers[name]; ok {
		panic(fmt.Errorf("codec: resolver called twice for adapter %s", name))
	}
	codecResolvers[name] = resolver
}

func newCodec(codecType CodecType) (r Codec, err error) {
	resolver, ok := codecResolvers[codecType]
	if !ok {
		return nil, fmt.Errorf("codec: unknown adapter name %q (forgotten import?)", codecType)
	}

	key := fmt.Sprintf("%s", codecType)

	_, value, err := codecMap.SetIfAbsentCb(key, func(input ...interface{}) (interface{}, error) {
		rsv := input[0].(Resolver)
		return rsv.Resolve(codecType)
	}, resolver)
	if err != nil {
		return
	}
	r = value.(Codec)
	return
}

func GetCodec(codecType CodecType) (Codec, error) {

	return newCodec(codecType)

}
