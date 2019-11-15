package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	_ "github.com/sereiner/duo/codec/gob"
	_ "github.com/sereiner/duo/codec/msgpack"
	"github.com/sereiner/duo/component"

	"github.com/sereiner/duo/codec"
	"github.com/sereiner/duo/context"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var typeOfContext = reflect.TypeOf((*context.IContext)(nil)).Elem()

type RPCServer interface {
	Register(rcvr interface{})
	Serve(network string, addr string) error
	Close() error
}

type Server struct {
	ln         net.Listener
	codec      codec.Codec
	serviceMap sync.Map
	mutex      sync.Mutex
	shutdown   bool
	c          component.IContainer
	*option
}

type service struct {
	name    string
	typ     reflect.Type
	rcvr    reflect.Value
	methods map[string]*methodType
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

func NewServer(c component.IContainer, opts ...Option) RPCServer {
	s := &Server{
		c:      c,
		option: &option{},
	}

	for _, op := range opts {
		op(s.option)
	}
	s.codecType = codec.MsgPackCodecType

	s.setCodec()
	return s
}

func (s *Server) setCodec() {
	code, err := codec.GetCodec(s.codecType)
	if err != nil {
		panic(err)
	}
	s.codec = code
}

func (s *Server) Register(rFunc interface{}) {

	rFuncType := reflect.TypeOf(rFunc)
	rFuncValue := reflect.ValueOf(rFunc)
	if rFuncType.Kind() != reflect.Func {
		panic("服务注册需要函数类型")
	}

	var rvalue []reflect.Value
	if rFuncType.NumIn() == 1 {
		if rFuncType.In(0).Name() != "IContainer" {
			panic("注册函数参数类型错误")
		}
		ivalue := make([]reflect.Value, 0, 1)
		ivalue = append(ivalue, reflect.ValueOf(s.c))
		rvalue = rFuncValue.Call(ivalue)
	} else {
		panic("注册函数参数个数错误")
	}

	if len(rvalue) != 1 {
		panic("类型错误,返回值只能有1个")
	}

	typ := rvalue[0].Type()
	name := getServiceName(rFuncType.Out(0).String())
	srv := new(service)
	srv.name = name
	srv.rcvr = rvalue[0]
	srv.typ = typ
	methods := suitableMethods(typ)
	srv.methods = methods

	if len(srv.methods) == 0 {
		var errorStr string
		method := suitableMethods(reflect.PtrTo(srv.typ))
		if len(method) != 0 {
			errorStr = "rpc.Register: type " + name + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			errorStr = "rpc.Register: type " + name + " has no exported methods of suitable type"
		}

		panic(errorStr)
	}

	if _, duplicate := s.serviceMap.LoadOrStore(name, srv); duplicate {
		panic("rpc: service already defined: " + name)
	}

}

func suitableMethods(typ reflect.Type) map[string]*methodType {

	methods := make(map[string]*methodType)

	for m := 0; m < typ.NumMethod(); m++ {

		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name

		// 方法必须是可导出的
		if method.PkgPath != "" {
			continue
		}
		// 需要有三个参数: receiver, *context.Context, req
		if mtype.NumIn() != 3 {

			log.Println("method", mname, "has wrong number of ins:", mtype.NumIn())
			continue
		}
		// 第一个参数必须是 *context.Context
		ctxType := mtype.In(1)
		if !ctxType.Implements(typeOfContext) {
			log.Println("method", mname, " must use context.Context as the first parameter")
			continue
		}

		// 第二个参数是请求参数
		argType := mtype.In(2)
		if !isExportedOrBuiltinType(argType) {
			log.Println(mname, "parameter type not exported:", argType)
			continue
		}

		// 必须有两个个返回值
		if mtype.NumOut() != 2 {
			log.Println("method", mname, "has wrong number of outs:", mtype.NumOut())
			continue
		}

		// 第一个返回值是结果，必须是指针类型的
		replyType := mtype.Out(0)
		if replyType.Kind() != reflect.Ptr {
			log.Println("method", mname, "reply type not a pointer:", replyType)
			continue
		}
		// 返回值的类型必须是可导出的
		if !isExportedOrBuiltinType(replyType) {
			log.Println("method", mname, "reply type not exported:", replyType)
			continue
		}

		// 第二个返回值类型必须是error
		if returnType := mtype.Out(1); returnType != typeOfError {
			log.Println("method", mname, "returns", returnType.String(), "not error")
			continue
		}

		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}

	return methods

}

func getServiceName(s string) string {

	if len(s) == 0 {
		return ""
	}

	return strings.Trim(s, "*")

}

func isExportedOrBuiltinType(t reflect.Type) bool {

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return isExported(t.Name()) || t.PkgPath() == ""
}

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

func (s *Server) Serve(network string, addr string) error {

	ln, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	s.ln = ln

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return err
		}
		go s.serveTransport(conn)
	}

}

func (s *Server) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.shutdown = true

	err := s.ln.Close()

	s.serviceMap.Range(func(key, value interface{}) bool {
		s.serviceMap.Delete(key)
		return true
	})
	return err
}

type Request struct {
	Seq   uint32
	Reply interface{}
	Data  []byte
}

func (s *Server) serveTransport(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("client has closed this connection: %s", conn.RemoteAddr().String())
			} else if strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("rpc: connection %s is closed", conn.RemoteAddr().String())
			} else {
				log.Printf("rpc: failed to read request: %v", err)
			}
			return
		}

		requestMsg := context.GetMessage()
		err = s.codec.Decode(buf[:n], requestMsg)
		if err != nil {
			log.Println(err)
			return
		}

		sname := requestMsg.ServiceName
		mname := requestMsg.MethodName

		srvInterface, ok := s.serviceMap.Load(sname)
		if !ok {
			s.writeErrorResponse(requestMsg, conn, "can not find service")
			return
		}

		srv, ok := srvInterface.(*service)
		if !ok {
			s.writeErrorResponse(requestMsg, conn, "not *service type")
			return
		}

		mtype, ok := srv.methods[mname]
		if !ok {
			s.writeErrorResponse(requestMsg, conn, "can not find method")
			return
		}

		argv := newValue(mtype.ArgType)
		ctx := context.GetContext()
		ctx.SetSeq(requestMsg.Seq)

		err = s.codec.Decode(requestMsg.Data, argv)
		if err != nil {
			log.Println(err)
			return
		}
		go s.call(conn, mtype, requestMsg, srv, ctx, argv)

	}
}
func (s *Server) call(conn net.Conn, mtype *methodType, requestMsg *context.Message, srv *service, ctx *context.Context, argv interface{}) {

	defer func() {
		if err := recover(); err != nil {
			s.writeErrorResponse(requestMsg, conn, fmt.Errorf("%v", err).Error())
		}
	}()

	var returns []reflect.Value
	if mtype.ArgType.Kind() != reflect.Ptr {
		returns = mtype.method.Func.Call([]reflect.Value{srv.rcvr,
			reflect.ValueOf(ctx),
			reflect.ValueOf(argv).Elem(),
		})
	} else {
		returns = mtype.method.Func.Call([]reflect.Value{srv.rcvr,
			reflect.ValueOf(ctx),
			reflect.ValueOf(argv),
		})
	}

	if len(returns) != 2 || returns[1].Interface() != nil {
		err := returns[1].Interface().(error)
		s.writeErrorResponse(requestMsg, conn, err.Error())
		return
	}

	reBt, err := s.codec.Encode(returns[0].Interface())
	if err != nil {
		log.Println(err)
		return
	}

	requestMsg.Data = reBt
	res, err := s.codec.Encode(requestMsg)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = conn.Write(res)
	if err != nil {
		log.Println(err)
		return
	}

	ctx.Close()
}

func newValue(t reflect.Type) interface{} {

	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	} else {
		return reflect.New(t).Interface()
	}

}

func (s *Server) writeErrorResponse(response *context.Message, w io.Writer, err string) {

	response.Error = err
	response.Data = nil
	bt, _ := s.codec.Encode(response)
	_, _ = w.Write(bt)

}
