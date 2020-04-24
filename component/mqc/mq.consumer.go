package mqc

import (
	"fmt"
	"github.com/sereiner/duo/component/mqc/conf"
)

type MQConsumer interface {
	Connect(topic string, channel string, address string) (err error)
	Consume(concurrency int, callback func(string)) (err error)
	UnConsume()
	Close() (err error)
}

func GetConsumer(address string, opts ...Option) (MQConsumer, error) {
	//	获取队列名
	mqType, raddr, err := conf.GetMqName(address)
	if err != nil {
		return nil, err
	}
	// 获取适配器
	adapter, ok := mqConsumerAdapters[mqType]
	if !ok {
		return nil, fmt.Errorf("该类型的mq消费者适配器未注入,mqType:%s", mqType)
	}
	// 构建消费者
	return adapter.Resolve(raddr[0], opts...)
}
