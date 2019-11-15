package client

import (
	"log"

	"github.com/sereiner/duo/context"
)

type CallFunc func(ctx *context.Context, serviceMethod string, arg interface{}) (reply []byte, err error)

type MiddlewareFunc func(CallFunc) CallFunc

func WrapLog(g CallFunc) CallFunc {
	return func(ctx *context.Context, serviceMethod string, arg interface{}) (reply []byte, err error) {

		log.Println("---------开始-------")

		reply, err = g(ctx, serviceMethod, arg)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("---------结束-------")

		return
	}
}
