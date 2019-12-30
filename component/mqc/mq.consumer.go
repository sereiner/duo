package mqc

import (
	"fmt"
	"github.com/sereiner/duo/component/mqc/util"
)

type MQConsumer interface {
	Connect() (err error)
	Consume(queue string, concurrency int, callback func(IMessage)) (err error)
	UnConsume(queue string)
	Close() (err error)
}

func GetConsumer(address string, opts ...Option) (MQConsumer, error) {
	//	获取队列名
	mqType, raddr, err := util.GetMqName(address)
	if err != nil {
		return nil, err
	}
	// 获取适配器
	adapter, ok := mqConsumerAdapters[mqType]
	if !ok {
		return nil, fmt.Errorf("该类型的mq消费者适配器未注入,mqType:%s", mqType)
	}
	adapter.Resolve(raddr[0], opts...)
	return nil, nil
}
