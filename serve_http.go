package wildcat

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type adaptServeHTTP struct {
	h http.Handler
}

func AdaptServeHTTP(h http.Handler) Handler {
	return &adaptServeHTTP{h}
}

func (a *adaptServeHTTP) convertHeader(hp *HTTPParser) http.Header {
	header := make(http.Header)

	for _, h := range hp.Headers {
		if h.Name == nil {
			continue
		}

		header[string(h.Name)] = append(header[string(h.Name)], string(h.Value))
	}

	return header
}

func (a *adaptServeHTTP) HandleConnection(hp *HTTPParser, rest []byte, c net.Conn) {
	u, err := url.Parse(fmt.Sprintf("http://%s/%s", string(hp.Host()), string(hp.Path)))
	if err != nil {
		log.Fatal(err)
		return
	}

	var protoMajor int
	var protoMinor int

	switch string(hp.Version) {
	case "HTTP/0.9":
		protoMinor = 9
	case "HTTP/1.0":
		protoMajor = 1
	case "HTTP/1.1":
		protoMajor = 1
		protoMinor = 1
	}

	req := http.Request{
		Method:        string(hp.Method),
		URL:           u,
		Proto:         string(hp.Version),
		ProtoMajor:    protoMajor,
		ProtoMinor:    protoMinor,
		Header:        a.convertHeader(hp),
		Body:          hp.BodyReader(rest, c),
		ContentLength: hp.ContentLength(),
		Host:          string(hp.Host()),
		RequestURI:    string(hp.Path),
		RemoteAddr:    c.RemoteAddr().String(),
	}

	w := &responseWriter{c: c}
	w.init()

	a.h.ServeHTTP(w, &req)

	c.Close()
}

type responseWriter struct {
	c net.Conn

	code   int
	header http.Header

	wroteHeader bool
}

func (r *responseWriter) init() {
	r.code = 200
	r.header = make(http.Header)
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}

	var buf bytes.Buffer

	for k, v := range r.header {
		buf.WriteString(k)
		buf.WriteString(": ")

		if len(v) == 1 {
			buf.WriteString(v[0])
		} else {
			buf.WriteString(strings.Join(v, ", "))
		}

		buf.WriteString("\r\n")
	}

	r.c.Write(buf.Bytes())

	r.wroteHeader = true
}

func (r *responseWriter) Write(buf []byte) (int, error) {
	r.WriteHeader(r.code)
	return r.c.Write(buf)
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.c, bufio.NewReadWriter(bufio.NewReader(r.c), bufio.NewWriter(r.c)), nil
}
