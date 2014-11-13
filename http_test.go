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

	n, err := hp.Parse(simple)
	require.NoError(t, err)

	assert.Equal(t, n, len(simple))

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

	_, err := hp.Parse(simpleHeaders)
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

	_, err := hp.Parse(simple3Headers)
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

func TestParseMissingData(t *testing.T) {
	hp := NewHTTPParser()

	_, err := hp.Parse(short)

	assert.Equal(t, err, ErrMissingData)
}

var multiline = []byte("GET / HTTP/1.0\r\nHost: cookie.com\r\n  more host\r\n\r\n")

func TestParseMultlineHeader(t *testing.T) {
	hp := NewHTTPParser()

	_, err := hp.Parse(multiline)
	require.NoError(t, err)

	assert.Equal(t, []byte("cookie.com more host"), hp.FindHeader([]byte("Host")))
}

var specialHeaders = []byte("GET / HTTP/1.0\r\nHost: cookie.com\r\nContent-Length: 50\r\n\r\n")

func TestParseSpecialHeaders(t *testing.T) {
	hp := NewHTTPParser()

	_, err := hp.Parse(specialHeaders)
	require.NoError(t, err)

	assert.Equal(t, []byte("cookie.com"), hp.Host())
	assert.Equal(t, 50, hp.ContentLength())
}

func TestFindHeaderIgnoresCase(t *testing.T) {
	hp := NewHTTPParser()

	_, err := hp.Parse(specialHeaders)
	require.NoError(t, err)

	assert.Equal(t, []byte("50"), hp.FindHeader([]byte("content-length")))
}

var multipleHeaders = []byte("GET / HTTP/1.0\r\nBar: foo\r\nBar: quz\r\n\r\n")

func TestFindAllHeaders(t *testing.T) {
	hp := NewHTTPParser()

	_, err := hp.Parse(multipleHeaders)
	require.NoError(t, err)

	bar := []byte("Bar")

	assert.Equal(t, []byte("foo"), hp.FindHeader(bar))
	assert.Equal(t, [][]byte{[]byte("foo"), []byte("quz")}, hp.FindAllHeaders(bar))
}
