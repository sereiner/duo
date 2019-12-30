package client

import (
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sereiner/duo/codec"
	_ "github.com/sereiner/duo/codec/gob"
	_ "github.com/sereiner/duo/codec/msgpack"
	"github.com/sereiner/duo/context"
)

var ErrorShutdown = errors.New("client is shut down")

type RPCClient interface {
	Go(ctx *context.Context, serviceMethod string, arg interface{}, reply []byte, done chan *Call) *Call
	Call(ctx *context.Context, serviceMethod string, arg interface{}) (reply []byte, err error)
	Decode(data []byte, value interface{}) error
	Close() error
}

type Call struct {
	ServiceMethod string
	Args          interface{}
	Reply         []byte
	Error         error
	Done          chan *Call
}

type Client struct {
	Codec        codec.Codec
	Conn         net.Conn
	pendingCalls sync.Map
	mutex        sync.Mutex
	shutdown     bool
	middleware   []MiddlewareFunc
	*option
}

func NewClient(network string, addr string, opts ...Option) (*Client, error) {

	client := &Client{
		middleware: []MiddlewareFunc{WrapLog},
	*option
}

// NewClient 构建一个客户端
func NewClient(network string, addr string, opts ...Option) (*Client, error) {

	// 1.构建客户端对象
	client := &Client{
		option: &option{
			codecType: codec.MsgPackCodecType,
		},
	}

	for _, op := range opts {
		op(client.option)
	}

	client.setCodec()

	// 2.设置配置参数
	for _, op := range opts {
		op(client.option)
	}
	// 3.设置传入的编码格式
	client.setCodec()

	// 4.连接服务端
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	client.Conn = conn

	// 5.保存连接
	client.Conn = conn

	// 6.
	go client.input()
	return client, nil
}

func (c *Call) done() {
	c.Done <- c
}

func (c *Client) Go(ctx *context.Context, serviceMethod string, args interface{}, reply []byte, done chan *Call) *Call {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply

	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done

	c.send(ctx, call)

	return call
}

func (c *Client) PreCall(ctx *context.Context, serviceMethod string, args interface{}) (reply []byte, err error) {
	Wrapper := func(goFunc CallFunc, middleware ...MiddlewareFunc) CallFunc {
		for i := len(middleware) - 1; i >= 0; i-- {
			goFunc = middleware[i](goFunc)
		}
		return goFunc
	}

	fn := Wrapper(c.Call, c.middleware...)
	return fn(ctx, serviceMethod, args)
}

func (c *Client) Call(ctx *context.Context, serviceMethod string, args interface{}) (reply []byte, err error) {

	var seq string
	if ctx.Value(context.RequestSeqKey) == nil {
		seq = context.GetSequence()
		ctx.WithValue(context.RequestSeqKey, seq)
	} else {
		seq = ctx.Value(context.RequestSeqKey).(string)
	}
func (c *Client) Call(ctx *context.Context, serviceMethod string, args interface{}) (reply []byte, err error) {

	seq := context.GetSequence()
	ctx.WithValue(context.RequestSeqKey, seq)

	canFn := func() {}
	if c.option.RequestTimeout != time.Duration(0) {
		canFn = ctx.WithTimeout(c.option.RequestTimeout)
		metaDataInterface := ctx.Value(context.MetaDataKey)
		var metaData map[string]interface{}
		if metaDataInterface == nil {
			metaData = make(map[string]interface{})
		} else {
			metaData = metaDataInterface.(map[string]interface{})
		}
		metaData[context.RequestTimeoutKey] = c.option.RequestTimeout.String()
		ctx.WithValue(context.MetaDataKey, metaData)
	}

	done := make(chan *Call, 1)

	call := c.Go(ctx, serviceMethod, args, reply, done)
	select {
	case <-ctx.Done():
		canFn()
		c.pendingCalls.Delete(seq)
		call.Error = errors.New("client request time out")
	case <-call.Done:
	}
	return call.Reply, call.Error
}

func (c *Client) Decode(data []byte, value interface{}) error {

	return c.Codec.Decode(data, value)

}

func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.shutdown = true

	c.pendingCalls.Range(func(key, value interface{}) bool {
		call, ok := value.(*Call)
		if ok {
			call.Error = ErrorShutdown
			call.done()
		}

		c.pendingCalls.Delete(key)
		return true
	})
	return nil
}

func (c *Client) send(ctx *context.Context, call *Call) {
	seq := ctx.Value(context.RequestSeqKey).(string)
	c.pendingCalls.Store(seq, call)
	msg := context.GetMessage()
	msg.Seq = seq
	msg.MethodName = strings.Split(call.ServiceMethod, "/")[1]
	msg.ServiceName = strings.Split(call.ServiceMethod, "/")[0]
	msg.MetaData = ctx.Value(context.MetaDataKey).(map[string]interface{})
	bt, err := c.Codec.Encode(call.Args)
	if err != nil {
		log.Println(err)
		c.pendingCalls.Delete(seq)
		call.Error = err
		call.done()
		msg.Close()
		return
	}
	msg.Data = bt

	data, err := c.Codec.Encode(msg)
	if err != nil {
		log.Println(err)
		c.pendingCalls.Delete(seq)
		call.Error = err
		call.done()
		msg.Close()
		return
	}

	_, err = c.Conn.Write(data)
	if err != nil {
		log.Println(err)
		c.pendingCalls.Delete(seq)
		call.Error = err
		call.done()
		msg.Close()
		return
	}
}

func (c *Client) input() {
	var err error
	var n int
	buf := make([]byte, 1024)

	for err == nil {

		// 读取连接得到的数据
		n, err = c.Conn.Read(buf)
		if err != nil {
			break
		}

		response := context.GetMessage()
		// 获取响应报文结构
		response := context.GetMessage()
		// 将消息解析成message结构
		err = c.Codec.Decode(buf[:n], response)
		if err != nil {
			break
		}

		seq := response.Seq
		callInterface, _ := c.pendingCalls.Load(seq)
		call := callInterface.(*Call)
		// 获取响应报文的请求序号
		seq := response.Seq
		// 不懂
		callInterface, _ := c.pendingCalls.Load(seq)
		call := callInterface.(*Call)
		// 删除请求序号
		c.pendingCalls.Delete(seq)
		switch {
		case call == nil:
			//请求已经被清理掉了，可能是已经超时了
		case response.Error != "":
			call.Error = errors.New(response.Error)
			call.done()
		default:
			// 不懂
			call.Reply = response.Data
			call.done()
		}
	}
	c.Close()
}

func (c *Client) setCodec() {
	code, err := codec.GetCodec(c.codecType)
	if err != nil {
		panic(err)
	}
	c.Codec = code
}
