package mqc

import (
	"fmt"
	"github.com/sereiner/duo/component/mqc/conf"
	"sync/atomic"
	"time"
)

type MQProducer interface {
	Connect() (err error)
	Publish(queue string, msg string, timeout time.Duration) (err error)
	ShutDown() (err error)
}

// 获取消息对象
func GetProducer(address string, opts ...Option) (mq MQProducer, err error) {
	// 获取消息中间件适配器
	mqType, addrs, err := conf.GetMqName(address)
	if err != nil {
		return
	}
	adapter, ok := mqcProducerAdapters[mqType]
	if !ok {
		err = fmt.Errorf("该mq适配器没有配置")
	}
	return adapter.Resolve(addrs[0], opts...)
}

// mqc消息对象
type ProcuderMessage struct {
	Headers   []string
	Queue     string
	Data      string
	SendTimes int32
	Timeout   time.Duration
}

// 记录发送次数
func (p *ProcuderMessage) AddSendTimes() {
	atomic.AddInt32(&p.SendTimes, 1)
}
