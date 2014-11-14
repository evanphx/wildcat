package main

import (
	"os"

	"github.com/evanphx/wildcat"
)

type Env int

func (_ Env) Redirect(hp *wildcat.HTTPParser) (string, string, error) {
	return "tcp", os.Getenv("REDIRECT"), nil
}

func main() {
	var e Env

	wildcat.ListenAndServe(":9594", wildcat.NewReverseProxy(e))

	// h := http.HandlerFunc(static)

	// http.ListenAndServe(":9594", h)

	// wildcat.ListenAndServe(":9594", wildcat.AdaptServeHTTP(h))

}
