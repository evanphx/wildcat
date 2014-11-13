package wildcat

import "net"

type Handler interface {
	HandleConnection(parser *HTTPParser, rest []byte, c net.Conn)
}

type Server struct {
	Listener net.Listener
	Handler  Handler
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

func (s *Server) handle(c net.Conn) {
	buf := make([]byte, OptimalBufferSize)

	hp := NewHTTPParser()

	for {
		n, err := c.Read(buf)
		if err != nil {
			return
		}

		res, err := hp.Parse(buf[:n])
		for err == ErrMissingData {
			var m int

			m, err = c.Read(buf[n:])
			if err != nil {
				return
			}

			n += m

			res, err = hp.Parse(buf[:n])
		}

		if err != nil {
			panic(err)
		}

		s.Handler.HandleConnection(hp, buf[res:n], c)
	}
}
