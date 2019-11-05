package main

import (
	"fmt"
	"github.com/sereiner/duo/codec/msgpack"
	"github.com/sereiner/duo/context"
	"net"
)

func main() {
	server, err := net.Listen("tcp", ":1208")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := server.Accept()
		if err != nil {

			continue
		}

		go Handler(conn)
	}

}

func Handler(conn net.Conn) {

	buf := make([]byte, 1024)
	codec := msgpack.MsgCodec{}

	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			fmt.Println(err)
			conn.Close()
			return
		}

		msg := context.GetMessage()
		err = codec.Decode(buf[:n],msg)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			return
		}



		fmt.Println(msg)
		msg.Data = map[string]interface{}{
			"a":1,
			"b":2,
		}

		data ,_ :=codec.Encode(msg)

		conn.Write(data)
	}
}
