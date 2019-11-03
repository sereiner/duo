package client

import (
	"github.com/sereiner/duo/context"
	"testing"
)

func TestClient(t *testing.T) {
	c,err :=NewClient("tcp","127.0.0.1:1208")
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	err = c.Call(context.NewContext(),"hello",map[string]interface{}{
		"name":"jack",
		"age":12,
	},map[string]interface{}{})
	if err != nil {
		t.Error(err)
		return
	}

}
