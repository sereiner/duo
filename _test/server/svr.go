package main

import (
	"github.com/sereiner/duo/component"
	"github.com/sereiner/duo/context"
	"github.com/sereiner/duo/server"
)

func main() {


	s := server.NewServer(component.New())
	s.Register(NewUserServer)
	s.Serve("tcp",":9999")
}





type UserServer struct {
	c component.IContainer
}

type Request struct {
	Name string
}

type Response struct {
	Name string
	Age int
}

func NewUserServer(c component.IContainer) *UserServer {
	return &UserServer{c:c}
}


func(u *UserServer) UserServer(ctx *context.Context,req *Request) (resp *Response,err error) {
	return &Response{
		Name: req.Name,
		Age:  20,
	},nil
}
