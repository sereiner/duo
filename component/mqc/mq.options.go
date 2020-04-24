package mqc

import (
	logger "github.com/sereiner/library/log"
	"time"
)

type OptionConf struct {
	Logger       logger.ILogger
	Address      string `json:"address"`
	Version      string `json:"version"`
	Persistent   string `json:"persistent"`
	Ack          string `json:"ack"`
	Retry        bool   `json:"retry"`
	Key          string `json:"key"`
	Raw          string `json:"raw"`
	QueueCount   int
	IntervalTime time.Duration
}

type Option func(*OptionConf)

// 是指直接根据配置对象来构建OptionConf吗
func WithConf(conf *OptionConf) Option {
	return func(o *OptionConf) {
		o = conf
	}
}

// 设置日志组件
func WithLogger(logger logger.ILogger) Option {
	return func(o *OptionConf) {
		o.Logger = logger
	}
}

//设置重连时间
func WithIntervalTime(time time.Duration) Option {
	return func(o *OptionConf) {
		o.IntervalTime = time
	}
}
