package msgpack

import (
	"github.com/sereiner/duo/codec"
	"github.com/vmihailenco/msgpack"
)

func init() {
	codec.Register(codec.MsgPackCodecType, &MsgCodec{})
}

type MsgCodec struct {
	name codec.CodecType
}

func (c *MsgCodec) Encode(value interface{}) ([]byte, error) {
	return msgpack.Marshal(value)
}

func (c *MsgCodec) Decode(data []byte, value interface{}) error {
	return msgpack.Unmarshal(data, value)
}

func (c *MsgCodec) Resolve(name codec.CodecType) (codec.Codec, error) {
	c.name = name
	return c, nil
}
