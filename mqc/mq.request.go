package mqc

import (
	"github.com/nsqio/go-nsq"

	"github.com/sereiner/duo/context"
)

type Request struct {
	ctx *context.Context
}

func New(ctx *context.Context) *Request {
	return &Request{
		ctx: ctx,
	}
}

func (h *Request) HandleMessage(msg *nsq.Message) error {
	// 还在想怎么存入消息
	return nil
}

func (h *Request) GetMsg() (string, error) {
	return "", nil
}
