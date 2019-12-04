package mqc

import (
	"github.com/nsqio/go-nsq"

	"github.com/sereiner/duo/context"
)

type Handler struct {
	ctx *context.Context
}

func New(ctx *context.Context) *Handler {
	return &Handler{
		ctx: ctx,
	}
}

func (h *Handler) HandleMessage(msg *nsq.Message) error {
	// 还在想怎么存入消息
	return nil
}

func (h *Handler) GetMsg() (string, error) {
	return "", nil
}
