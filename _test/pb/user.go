package pb

import (
	"github.com/sereiner/duo/_test/info"
	"github.com/sereiner/duo/context"
)

// IUserServer 用户接口
//type IUserServer interface {
//	// GetName 获取名称
//	GetName(ctx *context.Context, req *info.Request) (resp *info.Response, err error)
//
//	// GetName 获取名称2
//	GetName2(ctx *context.Context, req *info.Request) (resp *info.Response, err error)
//}

// IServer
type IServer interface {
	// GetAge 获取年龄
	GetAge(ctx *context.Context, req *info.Request) (resp *info.Response, err error)
}

//type IOrderServer interface {
//	GetOrder(ctx *context.Context, req *info.Request) (resp *info.Response, err error)
//}
