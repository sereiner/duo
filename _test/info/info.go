package info

import (
	"github.com/sereiner/duo/component"
	"github.com/sereiner/duo/context"
)

type UserServer struct {
	c component.IContainer
}

type Request struct {
	Name string
}

type Response struct {
	Name string
	Age  int
}

func NewUserServer(c component.IContainer) *UserServer {
	return &UserServer{c: c}
}

func (u *UserServer) GetName(ctx *context.Context, req *Request) (resp *Response, err error) {
	return &Response{
		Name: req.Name,
		Age:  20,
	}, nil
}
