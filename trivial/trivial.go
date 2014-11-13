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

	c.Write(static)
}

func main() {
	l, err := net.Listen("tcp", ":9594")
	if err != nil {
		panic(err)
	}

	s := &wildcat.Server{l, &X{}}

	s.Listen()
}
