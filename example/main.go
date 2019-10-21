package main

import (
	"context"
	"github.com/sereiner/duo/component"
	"github.com/sereiner/duo/duo"
	"github.com/sereiner/duo/example/rpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	app := duo.NewDouApp()
	app.Init(func(c component.IContainer) {
		rpc.RegisterSearchServiceServer(c.GetGrpcServer(), NewServer(c))
		reflection.Register(c.GetGrpcServer())
	})

	app.Run()
}

type server struct {
	c component.IContainer
}

func NewServer(c component.IContainer) *server {
	return &server{
		c: c,
	}
}

// SayHello implements helloworld.GreeterServer
func (s *server) Search(ctx context.Context, in *rpc.SearchRequest) (*rpc.SearchResponse, error) {
	log := s.c.Log(ctx)
	log.Info("--------搜索--------")
	log.Info("1. 开始")
	log.Info("2. 返回数据")
	return &rpc.SearchResponse{Response: "Hello " + in.Request}, nil

}
