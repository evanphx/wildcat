package wildcat

import (
	"bytes"
	"io"
	"net"
)

type Redirector interface {
	Redirect(hp *HTTPParser) (string, string, error)
}

type ReverseProxy struct {
	dir Redirector
}

func NewReverseProxy(dir Redirector) *ReverseProxy {
	return &ReverseProxy{dir}
}

func (r *ReverseProxy) HandleConnection(hp *HTTPParser, rest []byte, c net.Conn) {
	proto, where, err := r.dir.Redirect(hp)

	if err != nil {
		r.writeError(c, err)
		return
	}

	out, err := net.Dial(proto, where)
	if err != nil {
		r.writeError(c, err)
		return
	}

	err = r.writeHeader(hp, out)
	if err != nil {
		r.writeError(c, err)
		return
	}

	_, err = c.Write(rest)
	if err != nil {
		return
	}

	go io.Copy(c, out)

	io.Copy(out, c)
}

var (
	cError  = []byte("HTTP/1.0 500 Server Error\r\nContent-Length: 0\r\n\r\n")
	cSP     = []byte(" ")
	cHTTP11 = []byte("HTTP/1.1")
)

func (r *ReverseProxy) writeError(c net.Conn, err error) {
	c.Write(cError)
}

func (r *ReverseProxy) writeHeader(hp *HTTPParser, c net.Conn) error {
	var buf bytes.Buffer

	buf.Write(hp.Method)
	buf.Write(cSP)
	buf.Write(hp.Path)
	buf.Write(cSP)
	buf.Write(cHTTP11)
	buf.Write(cCRLF)

	for _, h := range hp.Headers {
		if h.Name == nil {
			continue
		}

		buf.Write(h.Name)
		buf.Write(cColon)
		buf.Write(h.Value)
		buf.Write(cCRLF)
	}

	buf.Write(cCRLF)

	_, err := c.Write(buf.Bytes())

	return err
}
