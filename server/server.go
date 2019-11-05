package server

import (
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/megaredfan/rpc-demo/protocol"
	"github.com/sereiner/duo/codec"
	"github.com/sereiner/duo/context"
)

type RPCServer interface {
	Register(rcvr interface{}, metaData map[string]string) error
	Serve(network string, addr string) error
	Close() error
}

type Server struct {
	ln         net.Listener
	codec      codec.Codec
	serviceMap sync.Map
	mutex      sync.Mutex
	shutdown   bool
	*option
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

type service struct {
	name    string
	typ     reflect.Type
	rcvr    reflect.Value
	methods map[string]*methodType
}

func NewServer(opts ...Option) RPCServer {
	s := new(Server)

	for _, op := range opts {
		op(s.option)
	}

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

func (s *Server) Register(rcvr interface{}, metaData map[string]string) error {
	typ := reflect.TypeOf(rcvr)
	name := typ.Name()
	srv := new(service)
	srv.name = name
	srv.rcvr = reflect.ValueOf(rcvr)
	srv.typ = typ
	methods := suitableMethods(typ, true)
	srv.methods = methods

	if len(srv.methods) == 0 {
		var errorStr string

		method := suitableMethods(reflect.PtrTo(srv.typ), false)
		if len(method) != 0 {
			errorStr = "rpcx.Register: type " + name + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			errorStr = "rpcx.Register: type " + name + " has no exported methods of suitable type"
		}
		log.Println(errorStr)
		return errors.New(errorStr)
	}
	if _, duplicate := s.serviceMap.LoadOrStore(name, srv); duplicate {
		return errors.New("rpc: service already defined: " + name)
	}
	return nil
}

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

//过滤符合规则的方法，从net.rpc包抄的
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name

		// 方法必须是可导出的
		if method.PkgPath != "" {
			continue
		}
		// 需要有四个参数: receiver, Context, args, *reply.
		if mtype.NumIn() != 4 {
			if reportErr {
				log.Println("method", mname, "has wrong number of ins:", mtype.NumIn())
			}
			continue
		}
		// 第一个参数必须是context.Context
		ctxType := mtype.In(1)
		if !ctxType.Implements(typeOfContext) {
			if reportErr {
				log.Println("method", mname, " must use context.Context as the first parameter")
			}
			continue
		}

		// 第二个参数是arg
		argType := mtype.In(2)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				log.Println(mname, "parameter type not exported:", argType)
			}
			continue
		}
		// 第三个参数是返回值，必须是指针类型的
		replyType := mtype.In(3)
		if replyType.Kind() != reflect.Ptr {
			if reportErr {
				log.Println("method", mname, "reply type not a pointer:", replyType)
			}
			continue
		}
		// 返回值的类型必须是可导出的
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				log.Println("method", mname, "reply type not exported:", replyType)
			}
			continue
		}
		// 必须有一个返回值
		if mtype.NumOut() != 1 {
			if reportErr {
				log.Println("method", mname, "has wrong number of outs:", mtype.NumOut())
			}
			continue
		}
		// 返回值类型必须是error
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				log.Println("method", mname, "returns", returnType.String(), "not error")
			}
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// Is this an exported - upper case - name?
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
				log.Printf("rpcx: connection %s is closed", conn.RemoteAddr().String())
			} else {
				log.Printf("rpcx: failed to read request: %v", err)
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
		replyv := newValue(mtype.ReplyType)

		ctx := context.NewContext()
		//err = s.codec.Decode(requestMsg.Data, argv)

		var returns []reflect.Value
		if mtype.ArgType.Kind() != reflect.Ptr {
			returns = mtype.method.Func.Call([]reflect.Value{srv.rcvr,
				reflect.ValueOf(ctx),
				reflect.ValueOf(argv).Elem(),
				reflect.ValueOf(replyv)})
		} else {
			returns = mtype.method.Func.Call([]reflect.Value{srv.rcvr,
				reflect.ValueOf(ctx),
				reflect.ValueOf(argv),
				reflect.ValueOf(replyv)})
		}
		if len(returns) > 0 && returns[0].Interface() != nil {
			err = returns[0].Interface().(error)
			s.writeErrorResponse(requestMsg, conn, err.Error())
			return
		}

		requestMsg.StatusCode = protocol.StatusOK
		requestMsg.Data = responseData

		_, err = conn.Write(protocol.EncodeMessage(s.option.ProtocolType, response))
		if err != nil {
			log.Println(err)
			return
		}
	}
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
	log.Println(response.Error)
	response.Data = nil
	bt, _ := s.codec.Encode(response)
	_, _ = w.Write(bt)
}
