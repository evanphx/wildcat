package wildcat

import (
	"crypto/tls"
	"net"
	"time"
)

type Handler interface {
	HandleConnection(parser *HTTPParser, rest []byte, c net.Conn)
}

type Server struct {
	Handler Handler
}

func (s *Server) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", ":9594")
	if err != nil {
		return err
	}

	return s.Serve(l)
}

func ListenAndServe(addr string, handler Handler) error {
	s := &Server{handler}

	return s.ListenAndServe(addr)
}

func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
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

// ListenAndServeTLS listens on the TCP network address srv.Addr and
// then calls Serve to handle requests on incoming TLS connections.
//
// Filenames containing a certificate and matching private key for
// the server must be provided. If the certificate is signed by a
// certificate authority, the certFile should be the concatenation
// of the server's certificate followed by the CA's certificate.
//
// If addr is blank, ":https" is used.
func (srv *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	if addr == "" {
		addr = ":https"
	}

	config := &tls.Config{}
	config.NextProtos = []string{"http/1.1"}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
	return srv.Serve(tlsListener)
}

func ListenAndServeTLS(addr string, certFile string, keyFile string, handler Handler) error {
	server := &Server{Handler: handler}
	return server.ListenAndServeTLS(addr, certFile, keyFile)
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}

	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)

	return tc, nil
}
