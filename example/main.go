package main

import (
	"google.golang.org/grpc"

	"context"

	"github.com/sereiner/duo/duo"

	"github.com/sereiner/duo/example/rpc"
)

func main() {
	app := duo.NewMicroApp(
		duo.WithName("app"),
		duo.WithSystemName("rpc"),
		duo.WithPlatName("rpc"))
	app.Initializing(func(s *grpc.Server) {
		rpc.RegisterSearchServiceServer(s, NewServer())
	})

	app.Start()
}

type server struct {
}

func NewServer() *server {
	return &server{}
}

func (s *server) Search(ctx context.Context, in *rpc.SearchRequest) (*rpc.SearchResponse, error) {

	return &rpc.SearchResponse{Response: "Hello " + in.Request}, nil

}