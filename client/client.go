package client

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/sereiner/duo/codec"
	_ "github.com/sereiner/duo/codec/gob"
	_ "github.com/sereiner/duo/codec/msgpack"
	"github.com/sereiner/duo/context"
)

var ErrorShutdown = errors.New("client is shut down")

type RPCClient interface {
	Go(ctx *context.Context, serviceMethod string, arg interface{}, reply interface{}, done chan *Call) *Call
	Call(ctx *context.Context, serviceMethod string, arg interface{}) (reply interface{}, err error)
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
	codec        codec.Codec
	Conn         net.Conn
	pendingCalls sync.Map
	mutex        sync.Mutex
	shutdown     bool
	*option
}

func NewClient(network string, addr string, opts ...Option) (*Client, error) {

	client := &Client{
		option: &option{
			codecType: codec.MsgPackCodecType,
		},
	}

	for _, op := range opts {
		op(client.option)
	}

	client.setCodec()

	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	client.Conn = conn

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
	msg.MethodName = call.ServiceMethod
	msg.ServiceName = call.ServiceMethod
	msg.MetaData = ctx.Value(context.MetaDataKey).(map[string]interface{})
	bt, err := c.codec.Encode(call.Args)
	if err != nil {
		log.Println(err)
		c.pendingCalls.Delete(seq)
		call.Error = err
		call.done()
		msg.Close()
		return
	}
	msg.Data = bt

	data, err := c.codec.Encode(msg)
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

		n, err = c.Conn.Read(buf)
		if err != nil {
			break
		}

		response := context.GetMessage()
		err = c.codec.Decode(buf[:n], response)
		if err != nil {
			break
		}

		seq := response.Seq
		callInterface, _ := c.pendingCalls.Load(seq)
		call := callInterface.(*Call)
		c.pendingCalls.Delete(seq)
		switch {
		case call == nil:
			//请求已经被清理掉了，可能是已经超时了
		case response.Error != "":
			call.Error = errors.New(response.Error)
			call.done()
		default:
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
	c.codec = code
}
