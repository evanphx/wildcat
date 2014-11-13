package main

import (
	"net"

	"github.com/evanphx/wildcat"
)

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:9594")
	if err != nil {
		panic(err)
	}

	s := &wildcat.Server{l}

	s.Listen()
}
