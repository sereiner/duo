package context

import (
	"context"
	"time"
)

const (
	RequestSeqKey      = "rpc_request_seq"
	RequestTimeoutKey  = "rpc_request_timeout"
	RequestDeadlineKey = "rpc_request_deadline"
	MetaDataKey        = "rpc_meta_data"
	AuthKey            = "rpc_auth"
	ProviderDegradeKey = "provider_degrade"
)

type IContext interface {
	context.Context
	WithValue(key, val interface{})
	WithTimeout(timeout time.Duration) context.CancelFunc
}



type Context struct {
	context.Context
}

func NewContext() *Context {
	return &Context{
		Context: context.Background(),
	}
}

func (c *Context) WithValue(key, val interface{}) {
	c.Context = context.WithValue(c.Context, key, val)
}

func (c *Context) WithTimeout(timeout time.Duration) context.CancelFunc {
	ctx, canFn := context.WithTimeout(c.Context, timeout)
	c.Context = ctx
	return canFn
}

func (c *Context) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}
