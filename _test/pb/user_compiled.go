// 根据接口生成的客户端和服务端的代码,需要自己实现服务端函数逻辑
package pb

import (
	"fmt"

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
	ctx.Log.Info("-----GetAge---")
	var age int
	switch req.Name {
	case "tom":
		age = 20
	case "jack":
		age = 21
	case "marry":
		age = 22
	case "jerry":
		age = 23

	}

	ctx.Log.Info("success")
	return &info.Response{
		Name: req.Name,
		Age:  age,
	}, fmt.Errorf("哈哈哈")

}

type ServerClient struct {
	cc *client.Client
}

func NewServerClient(c *client.Client) *ServerClient {
	return &ServerClient{cc: c}
}

// GetAge 获取年龄
func (c *ServerClient) GetAge(ctx *context.Context, req *info.Request) (resp *info.Response, err error) {

	reply, err := c.cc.PreCall(ctx, "pb.Server/GetAge", req)
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
