package mqc

import (
	"syscall"

	"github.com/judwhite/go-svc/svc"
	"github.com/nsqio/go-nsq"

	"github.com/sereiner/duo/context"
)

type Server interface {
	Run() error
}

type MqcServer struct {
	cfg       *nsq.Config
	topic     string
	address   string
	timeout   int
	channel   string
	request   *Request
	ctx       *context.Context
	nsqd      *NsqdServer
	nsqlookup *NsqLookUpServer
	nsqdadmin *NsqAdminServer
}

func NewMqcServer(ctx *context.Context) (mqc *MqcServer) {
	return &MqcServer{
		ctx:       ctx,
		nsqd:      NewNsqd(),
		nsqlookup: NewLookUp(),
		nsqdadmin: NewAdmin(),
		request:   New(ctx),
	}
}

// 启动三个守护进程组件
func (mqc *MqcServer) Run() error {
	if err := svc.Run(mqc.nsqlookup, syscall.SIGINT, syscall.SIGTERM); err != nil {
		//日志
		return err
	}
	if err := svc.Run(mqc.nsqd, syscall.SIGINT, syscall.SIGTERM); err != nil {
		//日志
		return err
	}
	if err := svc.Run(mqc.nsqdadmin, syscall.SIGINT, syscall.SIGTERM); err != nil {
		//日志
		return err
	}
	return nil
}

// 注册消费者
func (mqc *MqcServer) Config(topic, channel, address string, timeout int, i interface{}) error {
	// 初始化配置
	cfg := nsq.NewConfig()
	// 设置重连时间
	//报错
	//cfg.LookupdPollInterval =  timeout * time.Second
	mqc.cfg = cfg
	return nil
}

// Register注册消费者
func (mqc *MqcServer) Register(handle nsq.Handler) error {
	// 初始化消费者
	c, err := nsq.NewConsumer(mqc.topic, mqc.channel, mqc.cfg)
	if err != nil {
		return err
	}
	c.SetLogger(nil, 0)
	// 注册消费者消费方法
	c.AddHandler(handle)
	// 与注册nsq中心进行连接
	c.ConnectToNSQLookupd(mqc.address)

	select {
	case <-mqc.ctx.Done():

		return mqc.ctx.Err()
	}
	return nil
}

func (mqc *MqcServer) GetString() {

}
