package main

import (
	"github.com/sereiner/duo/_test/info"
	"github.com/sereiner/duo/component"
	"github.com/sereiner/duo/server"
)

func main() {

	s := server.NewServer(component.New())
	s.Register(info.NewUserServer)
	s.Serve("tcp", ":9999")
}
