package wildcat

import (
	"bufio"
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simple = []byte("GET / HTTP/1.0\r\n\r\n")

func TestParseSimple(t *testing.T) {
	hp := NewHTTPParser()

	err := hp.Parse(simple)
	require.NoError(t, err)

	assert.Equal(t, []byte("HTTP/1.0"), hp.Version)

	assert.Equal(t, []byte("/"), hp.Path)
	assert.Equal(t, []byte("GET"), hp.Method)
}

func BenchmarkParseSimple(b *testing.B) {
	hp := NewHTTPParser()

	for i := 0; i < b.N; i++ {
		hp.Parse(simple)
	}
}

func BenchmarkNetHTTP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := bufio.NewReader(bytes.NewReader(simple))
		http.ReadRequest(buf)
	}
}

var simpleHeaders = []byte("GET / HTTP/1.0\r\nHost: cookie.com\r\n\r\n")

func TestParseSimpleHeaders(t *testing.T) {
	hp := NewHTTPParser()

	err := hp.Parse(simpleHeaders)
	require.NoError(t, err)

	assert.Equal(t, []byte("cookie.com"), hp.FindHeader([]byte("Host")))
}

func BenchmarkParseSimpleHeaders(b *testing.B) {
	hp := NewHTTPParser()

	for i := 0; i < b.N; i++ {
		hp.Parse(simpleHeaders)
	}
}

var simple3Headers = []byte("GET / HTTP/1.0\r\nHost: cookie.com\r\nDate: foobar\r\nAccept: these/that\r\n\r\n")

func TestParseSimple3Headers(t *testing.T) {
	hp := NewHTTPParser()

	err := hp.Parse(simple3Headers)
	require.NoError(t, err)

	assert.Equal(t, []byte("cookie.com"), hp.FindHeader([]byte("Host")))
	assert.Equal(t, []byte("foobar"), hp.FindHeader([]byte("Date")))
	assert.Equal(t, []byte("these/that"), hp.FindHeader([]byte("Accept")))
}

func BenchmarkParseSimple3Headers(b *testing.B) {
	hp := NewHTTPParser()

	for i := 0; i < b.N; i++ {
		hp.Parse(simple3Headers)
	}
}

func BenchmarkNetHTTP3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := bufio.NewReader(bytes.NewReader(simple3Headers))
		http.ReadRequest(buf)
	}
}

var short = []byte("GET / HT")

func nTestParseMissingData(t *testing.T) {
	hp := NewHTTPParser()

	err := hp.Parse(short)

	assert.Equal(t, err, ErrMissingData)
}
