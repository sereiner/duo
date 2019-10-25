package component

import (
	"reflect"

	"google.golang.org/grpc"
)

var _ IServiceRegistry = &ServiceRegistry{}

type IComponentRegistry interface {
	GetServices() func(server *grpc.Server)

	IServiceRegistry
}

type IServiceRegistry interface {
	Initializing(c func(server *grpc.Server))
}

type ServiceRegistry struct {
	servicesFunc func(server *grpc.Server)
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{}
}

func (s *ServiceRegistry) isConstructor(h interface{}) bool {
	fv := reflect.ValueOf(h)
	tp := reflect.TypeOf(h)
	if fv.Kind() != reflect.Func || tp.NumIn() > 1 || tp.NumOut() > 2 || tp.NumOut() == 0 {
		return false
	}
	if tp.NumIn() == 1 && tp.In(0).Name() == "IContainer" {
		return true
	}
	if tp.NumIn() == 0 {
		return true
	}
	return false
}

func (s *ServiceRegistry) isHandler(h interface{}) bool {
	fv := reflect.ValueOf(h)
	tp := reflect.TypeOf(h)
	return fv.Kind() == reflect.Func && tp.NumIn() == 1 && tp.NumOut() == 1
}

func (s *ServiceRegistry) Initializing(c func(server *grpc.Server)) {
	s.servicesFunc = c
}

func (s *ServiceRegistry) GetServices() func(server *grpc.Server) {
	return s.servicesFunc
}
