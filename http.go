package wildcat

import (
	"bytes"
	"strconv"

	"github.com/vektra/errors"
)

type header struct {
	Name  []byte
	Value []byte
}

type HTTPParser struct {
	Method, Path, Version []byte

	headers      []header
	totalHeaders int

	host     []byte
	hostRead bool

	contentLength     int
	contentLengthRead bool
}

const DefaultHeaderSlice = 10

// Create a new parser
func NewHTTPParser() *HTTPParser {
	return NewSizedHTTPParser(DefaultHeaderSlice)
}

// Create a new parser allocating size for size headers
func NewSizedHTTPParser(size int) *HTTPParser {
	return &HTTPParser{
		headers:      make([]header, size),
		totalHeaders: size,
	}
}

var (
	ErrBadProto    = errors.New("bad protocol")
	ErrMissingData = errors.New("missing data")
	ErrUnsupported = errors.New("unsupported http feature")
)

// Parse the buffer as an HTTP Request. The buffer must contain the entire
// request or Parse will return ErrMissingData for the caller to get more
// data. (this thusly favors getting a completed request in a single Read()
// call).
//
// Returns the number of bytes used by the header (thus where the body begins).
// Also can return ErrUnsupported if an HTTP feature is detected but not supported.
func (hp *HTTPParser) Parse(input []byte) (int, error) {
	var headers int
	var path int
	var ok bool

	total := len(input)

method:
	for i := 0; i < total; i++ {
		switch input[i] {
		case ' ', '\t':
			hp.Method = input[0:i]
			ok = true
			path = i + 1
			break method
		}
	}

	if !ok {
		return 0, ErrMissingData
	}

	var version int

	ok = false

path:
	for i := path; i < total; i++ {
		switch input[i] {
		case ' ', '\t':
			ok = true
			hp.Path = input[path:i]
			version = i + 1
			break path
		}
	}

	if !ok {
		return 0, ErrMissingData
	}

	var state int

	ok = false
loop:
	for i := version; i < total; i++ {
		c := input[i]

		switch state {
		case 0:
			switch c {
			case '\r':
				hp.Version = input[version:i]
				state = 1
			case '\n':
				hp.Version = input[version:i]
				headers = i + 1
				ok = true
				break loop
			}
		case 1:
			if c != '\n' {
				return 0, errors.Context(ErrBadProto, "missing newline in version")
			}
			headers = i + 1
			ok = true
			break loop
		}
	}

	if !ok {
		return 0, ErrMissingData
	}

	var h int

	var headerName []byte

	state = 5

	start := headers

	for i := headers; i < total; i++ {
		switch state {
		case 5:
			switch input[i] {
			case '\r':
				state = 6
			case '\n':
				return i + 1, nil
			case ' ', '\t':
				state = 7
			default:
				start = i
				state = 0
			}
		case 6:
			if input[i] != '\n' {
				return 0, ErrBadProto
			}

			return i + 1, nil
		case 0:
			if input[i] == ':' {
				headerName = input[start:i]
				state = 1
			}
		case 1:
			switch input[i] {
			case ' ', '\t':
				continue
			}

			start = i
			state = 2
		case 2:
			switch input[i] {
			case '\r':
				state = 3
			case '\n':
				state = 5
			default:
				continue
			}

			hp.headers[h] = header{headerName, input[start:i]}
			h++

			if h == hp.totalHeaders {
				newHeaders := make([]header, hp.totalHeaders+10)
				copy(newHeaders, hp.headers)
				hp.headers = newHeaders
				hp.totalHeaders += 10
			}
		case 3:
			if input[i] != '\n' {
				return 0, ErrBadProto
			}
			state = 5

		case 7:
			switch input[i] {
			case ' ', '\t':
				continue
			}

			start = i
			state = 8
		case 8:
			switch input[i] {
			case '\r':
				state = 3
			case '\n':
				state = 5
			default:
				continue
			}

			cur := hp.headers[h-1].Value

			newheader := make([]byte, len(cur)+1+(i-start))
			copy(newheader, cur)
			copy(newheader[len(cur):], []byte(" "))
			copy(newheader[len(cur)+1:], input[start:i])

			hp.headers[h-1].Value = newheader
		}
	}

	return 0, ErrMissingData
}

// Return a value of a header matching name.
func (hp *HTTPParser) FindHeader(name []byte) []byte {
	for _, header := range hp.headers {
		if bytes.Equal(header.Name, name) {
			return header.Value
		}
	}

	for _, header := range hp.headers {
		if bytes.EqualFold(header.Name, name) {
			return header.Value
		}
	}

	return nil
}

// Return all values of a header matching name.
func (hp *HTTPParser) FindAllHeaders(name []byte) [][]byte {
	var headers [][]byte

	for _, header := range hp.headers {
		if bytes.EqualFold(header.Name, name) {
			headers = append(headers, header.Value)
		}
	}

	return headers
}

var cHost = []byte("Host")

// Return the value of the Host header
func (hp *HTTPParser) Host() []byte {
	if hp.hostRead {
		return hp.host
	}

	hp.hostRead = true
	hp.host = hp.FindHeader(cHost)
	return hp.host
}

var cContentLength = []byte("Content-Length")

// Return the value of the Content-Length header.
// A value of -1 indicates the header was not set.
func (hp *HTTPParser) ContentLength() int {
	if hp.contentLengthRead {
		return hp.contentLength
	}

	header := hp.FindHeader(cContentLength)
	if header != nil {
		i, err := strconv.Atoi(string(header))
		if err == nil {
			hp.contentLength = i
		}
	}

	hp.contentLengthRead = true
	return hp.contentLength
}
