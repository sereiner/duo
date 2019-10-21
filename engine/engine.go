package engine

import (
	"context"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sereiner/duo/component"
	logger "github.com/sereiner/library/log"
	"google.golang.org/grpc"
	"net"
	"time"
)

var (
	StRunning = "running"
	StStop    = "stop"
)

// 按数组下标依次执行拦截器,先进后出
var opts = []grpc.ServerOption{
	grpcmiddleware.WithUnaryServerChain(
		RecoveryInterceptor,
		GetClientIP,
		LoggingInterceptor,
	),
}

type ServiceEngine struct {
	server  *grpc.Server
	running string
	proto   string
	port    string
	addr    string
	host    string
}

func NewServiceEngine() *ServiceEngine {

	s := grpc.NewServer(opts...)
	return &ServiceEngine{
		server: s,
	}
}

func (s *ServiceEngine) Install(installFunc func(c component.IContainer)) {
	installFunc(s)
}

func (s *ServiceEngine) Log(ctx context.Context) *logger.Logger {
	v, ok := ctx.Value("__log__").(*logger.Logger)
	if ok {
		return v
	}
	return logger.GetSession("__log__", logger.CreateSession())
}

func (s *ServiceEngine) GetGrpcServer() *grpc.Server {
	if s.server == nil {
		s.server = grpc.NewServer(opts...)
	}

	return s.server
}

func (s *ServiceEngine) Start() error {

	s.proto = "tcp"
	s.running = StRunning
	if s.addr == "" {
		s.addr = ":9090"
	}

	errChan := make(chan error, 1)
	go func(ch chan error) {
		lis, err := net.Listen("tcp", s.addr)
		if err != nil {
			ch <- err
			return
		}

		if err := s.server.Serve(lis); err != nil {
			ch <- err
		}
	}(errChan)
	select {
	case <-time.After(time.Millisecond * 500):
		return nil
	case err := <-errChan:
		s.running = StStop
		return err
	}

}
