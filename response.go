package wildcat

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type Response struct {
	c net.Conn

	headers    []header
	numHeaders int
}

func NewResponse(c net.Conn) *Response {
	return &Response{
		c:          c,
		headers:    make([]header, 10),
		numHeaders: 0,
	}
}

func (r *Response) AddHeader(key, val []byte) {
	if r.numHeaders == len(r.headers) {
		newHeaders := make([]header, r.numHeaders+10)
		copy(newHeaders, r.headers)
		r.headers = newHeaders
	}

	r.headers[r.numHeaders] = header{key, val}
	r.numHeaders++
}

func (r *Response) AddStringHeader(key, val string) {
	r.AddHeader([]byte(key), []byte(val))
}

func (r *Response) WriteStatus(code int) {
	status := fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, statusText[code])
	r.c.Write([]byte(status))
}

var (
	cColon     = []byte(": ")
	cCRLF      = []byte("\r\n")
	cConnClose = []byte("Connection: close\r\n")
)

func (r *Response) WriteHeaders() {
	var buf bytes.Buffer

	for i := 0; i < r.numHeaders; i++ {
		h := &r.headers[i]
		buf.Write(h.Name)
		buf.Write(cColon)
		buf.Write(h.Value)
		buf.Write(cCRLF)
	}

	r.c.Write(buf.Bytes())
}

func (r *Response) WriteBodyBytes(body []byte) {
	r.c.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))))
	r.c.Write(body)
}

func (r *Response) WriteBodyString(body string) {
	r.WriteBodyBytes([]byte(body))
}

func (r *Response) WriteBodySizedStream(size int, reader io.Reader) {
	r.c.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", size)))
	io.Copy(r.c, reader)
}

func (r *Response) WriteBodyStream(size int, reader io.Reader) {
	r.c.Write(cConnClose)
	io.Copy(r.c, reader)
}
