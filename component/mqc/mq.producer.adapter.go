package mqc

import (
	"fmt"
)

var mqcProducerAdapters map[string]MqcProducerAdapter

// mqc服务适配器
type MqcProducerAdapter interface {
	Resolve(address string, opts ...Option) (MQProducer, error)
}

// 注册各mqc服务端适配器
func RegistMqcProducerAdapter(proto string, adapter MqcProducerAdapter) error {
	if adapter == nil {
		return fmt.Errorf("注入的生产者适配器为空,adapter:%v", adapter)
	}
	if _, ok := mqcProducerAdapters[proto]; ok {
		return fmt.Errorf("该生产者适配器已被注入,proto:%s,adapter:%v", proto, adapter)
	}
	mqcProducerAdapters[proto] = adapter
	return nil
}
