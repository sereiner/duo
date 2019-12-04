package mqc

import "github.com/nsqio/go-nsq"

type Producer struct {
	p *nsq.Producer
	//其他扩展
}

func New(address string) (*Producer, error) {
	producer, err := nsq.NewProducer(address, nsq.NewConfig())
	if err != nil {
		return nil, err
	}
	if err = producer.Ping(); err != nil {
		producer.Stop()
		producer = nil
	}
	return nil, err
}

func (obj *Producer) Push(topic, msg string) error {
	return obj.p.Publish(topic, []byte(msg))
}
