package mqc

import (
	"fmt"
	"strings"
)

var mqcProducerAdapters map[string]MqcProducerAdapter

// mqc服务适配器
type MqcProducerAdapter interface {
	Resolve(address string, opts ...Option) (MQProducer, error)
}

// 注册各mqc服务端适配器
func RegistMqcProducerResolver(proto string, resolver MqcProducerAdapter) error {
	if resolver == nil {
		return fmt.Errorf("适配器为空")
	}
	if _, ok := mqcProducerAdapters[proto]; ok {
		return fmt.Errorf("该适配器已存在")
	}
	mqcProducerAdapters[proto] = resolver
	return nil
}

// 获取mq名称
func GetMqName(address string) (proto string, raddr []string, err error) {
	addrs := strings.Split(address, "://")
	if len(addrs) == 0 || len(addrs) > 2 {
		err = fmt.Errorf("MQ地址配置错误%s，格式:stomp://192.168.0.1:61613", addrs)
	}
	if len(addrs[0]) == 0 {
		err = fmt.Errorf("MQ地址配置错误%s，格式:stomp://192.168.0.1:61613", addrs)
	}
	proto = addrs[0]
	if len(raddr) > 1 {
		raddr = strings.Split(addrs[1], ",")
	}
	return
}
