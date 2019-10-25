package duo

import "github.com/sereiner/duo/server"

type Option func(*option)

type option struct {
	//RegistryAddr       string
	// 应用名称
	Name string
	// 平台名称
	PlatName string
	// 系统名称
	SystemName string
	// 集群名称
	ClusterName string
	IsDebug     bool
	Trace       string
}

//WithPlatName 设置平台名称
func WithPlatName(platName string) Option {
	return func(o *option) {
		o.PlatName = platName
	}
}

//WithSystemName 设置系统名称
func WithSystemName(systemName string) Option {
	return func(o *option) {
		o.SystemName = systemName
	}
}

//WithClusterName 设置集群名称
func WithClusterName(clusterName string) Option {
	return func(o *option) {
		o.ClusterName = clusterName
	}
}

//WithName 设置系统全名
func WithName(name string) Option {
	return func(o *option) {
		o.Name = name
	}
}

//WithDebug 设置debug模式
func WithDebug() Option {
	server.IS_DEBUG = true
	return func(o *option) {
		o.IsDebug = true
	}
}
