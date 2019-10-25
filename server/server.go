package server

import (
	"net"
	"time"

	logger "github.com/sereiner/library/log"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
)

var (
	IS_DEBUG  = false
	StRunning = "running"
	StStop    = "stop"
)

type ServerEngine struct {
	*option
	engine      *grpc.Server
	ServiceFunc func(server *grpc.Server)
	running     string
	proto       string
	port        string
	addr        string
	host        string
}

func NewServiceEngine(name, address string, opts ...Option) *ServerEngine {
	s := &ServerEngine{
		option: &option{},
		addr:   address,
	}

	for _, opt := range opts {
		opt(s.option)
	}

	if s.Logger == nil {
		s.Logger = logger.GetSession(name, logger.CreateSession())
	}

	s.engine = grpc.NewServer(s.serverOption...)

	return s
}

// Run the http server
func (s *ServerEngine) Run() error {

	s.proto = "tcp"
	s.running = StRunning
	errChan := make(chan error, 1)
	go func(ch chan error) {
		lis, err := net.Listen("tcp", s.addr)
		if err != nil {
			ch <- err
			return
		}

		s.ServiceFunc(s.engine)
		reflection.Register(s.engine)
		if err := s.engine.Serve(lis); err != nil {
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

//Shutdown 关闭服务器
func (s *ServerEngine) Shutdown(timeout time.Duration) {
	if s.engine != nil {
		s.running = StStop
		s.engine.GracefulStop()
		time.Sleep(time.Second)

	}
}

//GetStatus 获取当前服务器状态
func (s *ServerEngine) GetStatus() string {
	return s.running
}
