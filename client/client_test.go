package client

import (
	"testing"
	"time"

	"github.com/sereiner/duo/context"
)

func TestClient(t *testing.T) {
	c, err := NewClient("tcp", "127.0.0.1:1208", WithRequestTimeout(time.Second*1))
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	reply, err := c.Call(context.NewContext(), "hello", map[string]interface{}{
		"name": "jack",
		"age":  12,
	})
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(reply)

}
