package server

import (
	logger "github.com/sereiner/library/log"
	"google.golang.org/grpc"
)

type option struct {
	ip string
	*logger.Logger
	showTrace    bool
	platName     string
	systemName   string
	clusterName  string
	serverType   string
	tls          []string
	serverOption []grpc.ServerOption
}

//Option 配置选项
type Option func(*option)

func WithName(platName string, systemName string, clusterName string, serverType string) Option {
	return func(o *option) {
		o.platName = platName
		o.systemName = systemName
		o.clusterName = clusterName
		o.serverType = serverType
	}
}

//WithLogger 设置日志记录组件
func WithLogger(logger *logger.Logger) Option {
	return func(o *option) {
		o.Logger = logger
	}
}

//WithShowTrace 显示跟踪信息
func WithShowTrace(b bool) Option {
	return func(o *option) {
		o.showTrace = b
	}
}

//WithIP 设置ip地址
func WithIP(ip string) Option {
	return func(o *option) {
		o.ip = ip
	}
}

//WithMetric 设置基于influxdb的系统监控组件
//func WithMetric(host string, dataBase string, userName string, password string, cron string) Option {
//	return func(o *option) {
//		o.metric.Restart(host, dataBase, userName, password, cron, o.Logger)
//	}
//}

//WithTLS 设置TLS证书(pem,key)
func WithTLS(tls []string) Option {
	return func(o *option) {
		if len(tls) == 2 {
			o.tls = tls
		}
	}
}

func WithServerOption(op []grpc.ServerOption) Option {
	return func(o *option) {
		o.serverOption = append(o.serverOption, op...)
	}
}
