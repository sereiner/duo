package context

import (
	"context"
	"sync"
	"time"

	logger "github.com/sereiner/library/log"
)

const (
	RequestSeqKey      = "rpc_request_seq"
	RequestTimeoutKey  = "rpc_request_timeout"
	RequestDeadlineKey = "rpc_request_deadline"
	MetaDataKey        = "rpc_meta_data"
	AuthKey            = "rpc_auth"
	ProviderDegradeKey = "provider_degrade"
)

var contextPool *sync.Pool

type IContext interface {
	context.Context
	WithValue(key, val interface{})
	WithTimeout(timeout time.Duration) context.CancelFunc
}

type Context struct {
	Log *logger.Logger
	context.Context
}

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return newContext()
		},
	}

}

func newContext() *Context {
	return &Context{
		Log:     logger.New("ctx"),
		Context: context.Background(),
	}
}

func GetContext() *Context {
	return contextPool.Get().(*Context)
}

func (c *Context) Close() {
	c.Context = context.Background()
	c.Log.Close()
	contextPool.Put(c)
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

func (c *Context) SetSeq(seqID string) {
	c.Log = logger.GetSession("ctx", seqID)
}
