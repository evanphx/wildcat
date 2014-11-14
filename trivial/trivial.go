package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/evanphx/wildcat"
)

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

func static(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello world\n"))
}

func main() {
	wildcat.ListenAndServe(":9594", &X{})

	// h := http.HandlerFunc(static)

	// http.ListenAndServe(":9594", h)

	// wildcat.ListenAndServe(":9594", wildcat.AdaptServeHTTP(h))

}
