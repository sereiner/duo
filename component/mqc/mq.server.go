package mqc

import (
	"fmt"
	dnsq "github.com/sereiner/duo/component/mqc/nsq"
	"github.com/sereiner/duo/context"
)

type MqcServer interface {
	Connect() (err error)
	Send(queue, message string) (err error)
	ShutDown()
}

// 获取消息对象
func GetQueue(ctx *context.Context, name, address string) (mq MqcServer, err error) {
	if address ==""{
		err = fmt.Errorf("服务地址不能为空！")
	}
	mq = dnsq.NewServer(ctx, address)
	
	return
}
