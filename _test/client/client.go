package main

import (
	"fmt"
	"github.com/sereiner/duo/_test/info"
	"github.com/sereiner/duo/_test/pb"
	"github.com/sereiner/duo/client"
	"github.com/sereiner/duo/context"
	"time"
)

func main() {
	c, err := client.NewClient("tcp", "127.0.0.1:9999", client.WithRequestTimeout(time.Second*1))
	if err != nil {
		panic(err)
	}
	defer c.Close()

	u := NewUserServerClient(c)
	fmt.Println(u.GetName(context.NewContext(), &info.Request{Name: "tom"}))
}

type UserServerClient struct {
	client.RPCClient
}

func NewUserServerClient(client client.RPCClient) pb.IUserServer {
	return &UserServerClient{
		RPCClient: client,
	}
}

func (u *UserServerClient) GetName(ctx *context.Context, req *info.Request) (resp *info.Response, err error) {

	reply, err := u.Call(ctx, "info.UserServer/GetName", req)
	if err != nil {
		panic(err)
	}

	m := &info.Response{}

	err = u.RPCClient.Decode(reply, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
