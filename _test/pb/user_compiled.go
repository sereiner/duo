// 根据接口生成的客户端和服务端的代码,需要自己实现服务端函数逻辑
package pb

import (
	"github.com/sereiner/duo/_test/info"
	"github.com/sereiner/duo/client"
	"github.com/sereiner/duo/component"
	"github.com/sereiner/duo/context"
)

type Server struct {
	c component.IContainer
}

func NewServer(c component.IContainer) *Server {
	return &Server{c: c}
}

// GetAge 获取年龄
func (a *Server) GetAge(ctx *context.Context, req *info.Request) (resp *info.Response, err error) {
	panic("server not implement GetAge")
}

type ServerClient struct {
	cc client.RPCClient
}

func NewServerClient(c client.RPCClient) *ServerClient {
	return &ServerClient{cc: c}
}

// GetAge 获取年龄
func (c *ServerClient) GetAge(ctx *context.Context, req *info.Request) (resp *info.Response, err error) {
	reply, err := c.cc.Call(ctx, "pb.Server/GetAge", req)
	if err != nil {
		return nil, err
	}
	m := &info.Response{}
	err = c.cc.Decode(reply, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
