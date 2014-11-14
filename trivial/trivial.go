package main

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/evanphx/wildcat"
)

var static = []byte("HTTP/1.1 200 OK\r\nContent-Length: 11\r\n\r\nhello world")

type X struct{}

func (x *X) HandleConnection(hp *wildcat.HTTPParser, rest []byte, c net.Conn) {
	cl := hp.ContentLength()

	fmt.Printf("host: %s\n", string(hp.Host()))
	fmt.Printf("content: %d\n", cl)

	if hp.Post() {
		body, err := ioutil.ReadAll(hp.BodyReader(rest, c))
		if err != nil {
			panic(err)
		}

		fmt.Printf("body: %s\n", body)
	}

	resp := wildcat.NewResponse(c)
	resp.AddStringHeader("X-Runtime", "8311323")

	resp.WriteStatus(200)
	resp.WriteHeaders()
	resp.WriteBodyBytes([]byte("hello world\n"))
}

func main() {
	wildcat.ListenAndServe(":9594", &X{})
}
