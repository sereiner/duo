package gob

import (
	"bytes"
	"encoding/gob"
	"github.com/sereiner/duo/codec"
)

func init() {
	codec.Register(codec.GobCodecType, &GobCodec{})
}

type GobCodec struct {
	name codec.CodecType
}

func (c *GobCodec) Encode(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(value)
	return buf.Bytes(), err
}

func (c *GobCodec) Decode(data []byte, value interface{}) error {

	buf := bytes.NewBuffer(data)
	return gob.NewDecoder(buf).Decode(value)

}

func (c *GobCodec) Resolve(name codec.CodecType) (codec.Codec, error) {
	c.name = name
	return c, nil
}
