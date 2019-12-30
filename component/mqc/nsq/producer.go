package nsqServer

import (
	"github.com/nsqio/go-nsq"
	"github.com/sereiner/duo/component/mqc"
	logger "github.com/sereiner/library/log"
	"time"
)

type NsqProducer struct {
	address string
	client  *nsq.Producer
	message chan *mqc.ProcuderMessage
	closeCh chan struct{} //怎么用
	done    bool
	*mqc.OptionConf
}

func NewNsqProducer(address string, opts ...mqc.Option) (producer *NsqProducer, err error) {
	producer = &NsqProducer{address: address,
		closeCh: make(chan struct{}),
	}
	producer.OptionConf = &mqc.OptionConf{
		Logger: logger.GetSession("mq.nsq", logger.CreateSession()),
	}
	for _, o := range opts {
		o(producer.OptionConf)
	}
	return
}

// 连接nsq服务
func (n *NsqProducer) Connect() (err error) {
	n.client, err = nsq.NewProducer(n.address, nsq.NewConfig())
	if err != nil {
		n.done = true
		n.OptionConf.Logger.Errorf("连接失败！,err:", err)
	}
	return n.client.Ping()
}

func (n *NsqProducer) Publish(queue string, msg string, timeout time.Duration) (err error) {
	if n.done {
		n.OptionConf.Logger.Errorf("连接已关闭")
		return
	}
	if msg == "" {
		n.OptionConf.Logger.Errorf("消息为空")
		return
	}
	return n.client.Publish(queue, []byte(msg))
}

func (n *NsqProducer) ShutDown() (err error) {
	if n.done {
		n.OptionConf.Logger.Errorf("队列已关闭")
		return
	}
	n.done = true
	n.client.Stop()
	n.client = nil
	return
}

func (n *NsqProducer) GetMessage() chan *mqc.ProcuderMessage {
	return n.message
}

//nsq 适配器
type nsqProducerAdapter struct {
}

// nsq适配器构建nsqproducer
func (adapter *nsqProducerAdapter) Resolve(address string, opts ...mqc.Option) (mqc.MQProducer, error) {
	return NewNsqProducer(address, opts...)
}

func init() {
	mqc.RegistMqcProducerAdapter("nsq", &nsqProducerAdapter{})
}
