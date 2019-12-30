package nsqServer

import (
	"fmt"
	"time"

	go_nsq "github.com/nsqio/go-nsq"

	"github.com/sereiner/library/concurrent/cmap"
	"github.com/sereiner/duo/context"
)

type NsqClient struct {
	ctx *context.Context
	address   string
	consumers cmap.ConcurrentMap
	closeCh  chan struct{}
}

type nsqComsumer struct {
	comsumer go_nsq.Consumer
	msgQueue chan *go_nsq.Message
	closeCh  chan struct{}
}

//处理消息
func (n *nsqComsumer) HandleMessage(msg *go_nsq.Message) error {
	fmt.Println("receive: ", msg.NSQDAddress, "message: ", string(msg.Body))
	n.msgQueue <- msg
	return nil
}

func NewNsqClient(ctx *context.Context,address string) (client *NsqClient,err error) {
	client := &NsqClient{ctx: ctx,address: address}
	client.closeCh = make(chan struct{})
	client.consumers = cmap.New(2)
	return 
}


// 如何与外界服务串联起来
func Init(ctx *context.Context, topic, channel, address string, intervale int) error {
	// c := NewConsumerT(ctx)
	cfg := go_nsq.NewConfig()
	if intervale == 0 {
		intervale = 15
	}
	cfg.LookupdPollInterval = time.Duration(intervale) * time.Second
	client, err := go_nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		panic(err)
	}
	//屏蔽系统日志
	client.SetLogger(nil, 0)
	// 添加消费者接口
	client.AddHandler(c)

	//建立NSQLookupd连接
	if err := client.ConnectToNSQLookupd(address); err != nil {
		return fmt.Errorf("连接nsq失败,err:", err)
	}
	ctx.Log.Infof("连接成功")
	return nil
}

func (n *NsqClient) Consume() (err error)  {
	
}

func (n *NsqClient) ShutDown()  {
	close(closeCh)
	n.consumers.IterCb(func(key string, value interface{}) bool {
		c := value.(*nsqConsumer)
		c.consumer.Stop()
		time.Sleep(time.Second)
		c.closeCh <- struct{}{}
		close(c.msgQueue)
		return true
	})
}
