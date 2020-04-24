package nsq

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/sereiner/duo/component/mqc"
)

type NsqConsumer struct {
	Address string
	//consumers cmap.ConcurrentMap//可否直接保存连接
	Titil   string
	Cilent  *nsq.Consumer
	CloseCh chan bool
	MsgCh   chan string
	Done    bool
	Cfg     *nsq.Config
	*mqc.OptionConf
}

func newNsqConsumer(address string, opts ...mqc.Option) (consumer *NsqConsumer, err error) {
	//构建配置
	for _, o := range opts {
		o(consumer.OptionConf)
	}
	//构建消费者对象
	consumer = &NsqConsumer{
		Address: address,
		Titil:   "消费者的唯一标识,没想好如何定义",
		CloseCh: make(chan bool),
		Cfg:     nsq.NewConfig(),
		MsgCh:   make(chan string),
	}

	return
}

func (n *NsqConsumer) Connect(topic string, channel string, address string) (err error) {
	//重连时间
	n.Cfg.LookupdPollInterval = n.OptionConf.IntervalTime
	//构建nsq连接
	n.Cilent, err = nsq.NewConsumer(topic, channel, n.Cfg) // 新建一个消费者
	if err != nil {
		n.Logger.Errorf("构建消费者异常:%v", err)
		return err
	}
	// 日志还没想好如何与外层定义
	n.Cilent.SetLogger(nil, 0)
	n.Cilent.AddHandler(n)
	return
}

func (n *NsqConsumer) HandleMessage(msg *nsq.Message) error {
	//得到消息
	fmt.Println("receive", msg.NSQDAddress, "message:", string(msg.Body))
	//存入消息
	for {
		select {
		case <-n.CloseCh:
			break
		}
		if string(msg.Body) != "" {
			n.MsgCh <- string(msg.Body)
		}
	}
	close(n.MsgCh)
	return nil
}

//消费消息
func (n *NsqConsumer) Consume(concurrency int, callback func(string)) (err error) {
	//检查是否关闭
	if n.Done {
		err = fmt.Errorf("消费者已关闭！")
		return
	}
	go func() {
		message := <-n.MsgCh
		callback(message)
	}()
	return
}

func (n *NsqConsumer) UnConsume() {
	n.CloseCh <- true
}

func (n *NsqConsumer) Close() (err error) {
	n.Done = true
	n.CloseCh <- true
	n.Cilent.Stop()
	return
}

type nsqConsumerAdapter struct {
}

func (adapter *nsqConsumerAdapter) Resolve(address string, opts ...mqc.Option) (mqc.MQConsumer, error) {
	return newNsqConsumer(address, opts...)
}

func init() {
	mqc.RegisteMqcConsumerAdapter("nsq", &nsqConsumerAdapter{})
}
