package context

import (
	"github.com/sereiner/library/utility"
)

func GetSequence() string {
	return utility.GetGUID()[0:9]
}


type Message struct {
	*Header
	Data []byte
}


type Header struct {
	Seq           string                 //序号, 用来唯一标识请求或响应
	StatusCode    int             		 //状态类型，用来标识一个请求是正常还是异常
	ServiceName   string                 //服务名
	MethodName    string                 //方法名
	Error         string                 //方法调用发生的异常
	MetaData      map[string]interface{} //其他元数据
}
