package mqc

import "fmt"

var mqConsumerAdapters map[string]MQConsumerAdapter

type MQConsumerAdapter interface {
	Resolve(address string, opts ...Option) (MQConsumer, error)
}

func RegisteMqcConsumerAdapter(proto string, adapter MQConsumerAdapter) error {
	if adapter == nil {
		return fmt.Errorf("注入的消费者适配器为空,adapter:%v", adapter)
	}
	if _, ok := mqConsumerAdapters[proto]; ok {
		return fmt.Errorf("该消费者适配器已被注入,proto:%s,adapter:%v", proto, adapter)
	}
	mqConsumerAdapters[proto] = adapter
	return nil
}
