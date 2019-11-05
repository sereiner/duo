package context

import (
	"fmt"
	"sync"

	"github.com/sereiner/library/utility"
)

func GetSequence() string {
	return utility.GetGUID()[0:9]
}

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return New()
		},
	}

}

type Header struct {
	Seq         string                 //序号, 用来唯一标识请求或响应
	StatusCode  int                    //状态类型，用来标识一个请求是正常还是异常
	ServiceName string                 //服务名
	MethodName  string                 //方法名
	Error       string                 //方法调用发生的异常
	MetaData    map[string]interface{} //其他元数据
}

type Message struct {
	*Header
	Data interface{}
}

func New() *Message {
	return &Message{
		Header: &Header{},
	}
}

func GetMessage() (m *Message) {
	return contextPool.Get().(*Message)
}

func (m *Message) String() string {
	return fmt.Sprintf("seq:%s ServiceName:%s MethodName:%s StatusCode:%d Error:%s MetaData:%v data:%s",
		m.Seq, m.ServiceName, m.MethodName, m.StatusCode, m.Error, m.MetaData, m.Data)
}

func (m *Message) Close() {

	m.Seq = ""
	m.StatusCode = 0
	m.ServiceName = ""
	m.MethodName = ""
	m.Error = ""
	m.MetaData = nil

	contextPool.Put(m)
}
