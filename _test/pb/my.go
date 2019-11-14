package pb

import (
	"github.com/sereiner/duo/_test/info"
	"github.com/sereiner/duo/context"

	"github.com/sereiner/duo/component"
)

// IUserServer 用户接口
type UserServer struct {
	c component.IContainer
}

// GetName 获取名称
func (a *UserServer) GetName(ctx *context.Context, req *info.Request) (resp *info.Response, err error) {
	panic("not implement")
}

// GetName 获取名称2
func (a *UserServer) GetName2(ctx *context.Context, req *info.Request) (resp *info.Response, err error) {
	panic("not implement")
}

// IServer
type Server struct {
	c component.IContainer
}

// GetAge 获取年龄
func (a *Server) GetAge(ctx *context.Context, req *info.Request) (resp *info.Response, err error) {
	panic("not implement")
}
