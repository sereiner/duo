package main

import (
	"github.com/sereiner/duo/_test/pb"
	"github.com/sereiner/duo/component"
	"github.com/sereiner/duo/server"
)

func main() {

	s := server.NewServer(component.New())
	s.Register(pb.NewServer)
	s.Serve("tcp", ":9999")
}
