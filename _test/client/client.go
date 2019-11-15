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

	u := pb.NewServerClient(c)
	fmt.Println(u.GetAge(context.NewContext(), &info.Request{Name: "tom"}))
}
