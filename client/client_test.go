package client

import (
	"testing"
	"time"

	"github.com/sereiner/duo/context"
)

type Request struct {
	Name string
}

func TestClient(t *testing.T) {
	c, err := NewClient("tcp", "127.0.0.1:9999", WithRequestTimeout(time.Second*1))
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	// 调用方法
	reply, err := c.Call(context.NewContext(), "UserServer", &Request{Name: "jack"})
	if err != nil {
		t.Error(err)
		return
	}

	m := map[string]interface{}{}
	// 解析返回值
	c.codec.Decode(reply, &m)

	t.Log(m)

}
