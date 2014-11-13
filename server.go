package wildcat

import (
	"fmt"
	"net"
)

type Server struct {
	Listener net.Listener
}

func (s *Server) Listen() error {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return err
		}

		go s.handle(conn)
	}
}

var static = []byte("HTTP/1.1 200 OK\r\nContent-Length: 11\r\n\r\nhello world")

func (s *Server) handle(c net.Conn) {
	fmt.Printf("new conn!\n")
	buf := make([]byte, 1500)

	hp := NewHTTPParser()

	for {
		n, err := c.Read(buf)
		if err != nil {
			return
		}

		_, err = hp.Parse(buf[0:n])
		if err != nil {
			panic(err)
		}

		c.Write(static)
	}
}
