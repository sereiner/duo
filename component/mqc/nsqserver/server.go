package nsqServer

import (
	"fmt"

	go_nsq "github.com/nsqio/go-nsq"

	"github.com/sereiner/duo/context"
)

type NsqServer struct {
	address  string
	ctx      *context.Context
	producer *go_nsq.Producer
	closeCh  chan struct{}
	done     bool
}

// 构建nsqserver
func NewServer(ctx *context.Context, address string) *NsqServer {
	nsqServer := &NsqServer{ctx: ctx, address: address}
	nsqServer.closeCh = make(chan struct{})
	return nsqServer
}

// Connect 初始化服务
func (n *NsqServer) Connect() (err error) {
	// 初始化生产者
	n.producer, err = go_nsq.NewProducer(n.address, go_nsq.NewConfig())
	if err != nil {
		err = fmt.Errorf("初始化服务失败，err:", err)
		return
	}
	// 测试连接
	if err = n.producer.Ping(); err != nil {
		n.ShutDown()
		err = fmt.Errorf("服务连接Nsqloogup失败，err:", err)
		return
	}
	n.ctx.Log.Info("ping nsq success")
	return
}

// 发送消息
func (n *NsqServer) Send(queue, message string) (err error) {
	if n.done {
		return fmt.Errorf("连接已关闭!")
	}
	if message == "" {
		err = fmt.Errorf("消息不能为空！")
		return
	}
	if err = n.producer.Publish(queue, []byte(message)); err != nil {
		err = fmt.Errorf("发送消息错误,err:", err)
		return
	}
	n.ctx.Log.Info("publish success")
	return
}

// 关闭连接
func (n *NsqServer) ShutDown() error {
	if n.done {
		return fmt.Errorf("连接已关闭!")
	}
	n.done = true
	close(n.closeCh)
	n.producer.Stop()
	return nil
}
