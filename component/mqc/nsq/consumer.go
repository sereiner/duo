package nsqServer

import (
	"github.com/sereiner/duo/component/mqc"
	"github.com/sereiner/library/concurrent/cmap"
)

type NsqConsumer struct {
	address   string
	consumers cmap.ConcurrentMap
	closeCh   chan struct{}
	*mqc.OptionConf
}

func NewNsqConsumer(address string, opts ...mqc.Option) (consumer *NsqConsumer, err error) {
	consumer = &NsqConsumer{
		address:   address,
		closeCh:   make(chan struct{}),
		consumers: cmap.New(2),
	}
	for _, o := range opts {
		o(consumer.OptionConf)
	}
	return
}

func (n *NsqConsumer) Connect() (err error) {
	return
}

func (n *NsqConsumer) Consume(queue string, concurrency int, callback func(mqc.IMessage)) (err error) {
	return
}

func (n *NsqConsumer) UnConsume(queue string) {

}

func (n *NsqConsumer) Close() (err error) {
	return
}

type nsqConsumerAdapter struct {
}

func (adapter *nsqConsumerAdapter) Resolve(address string, opts ...mqc.Option) (mqc.MQConsumer, error) {
	return NewNsqConsumer(address, opts...)
}

func init() {
	mqc.RegisteMqcConsumerAdapter("nsq", &nsqConsumerAdapter{})
}
