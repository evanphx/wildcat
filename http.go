package wildcat

import (
	"bytes"

	"github.com/vektra/errors"
)

type header struct {
	Name  []byte
	Value []byte
}

type HTTPParser struct {
	Method, Path, Version []byte

	headers []header
}

func NewHTTPParser() *HTTPParser {
	return &HTTPParser{
		headers: make([]header, 10, 10),
	}
}

const (
	eStart int = iota
	eVerb
	ePathStart
	ePath
	eVersionStart
	eVersionH
	eVersionHT
	eVersionHTT
	eVersionHTTP
	eVersionNumberStart
	eVersionNumberAll
	eVersionNumberAllN
	eHeaderStart
	eHeaderColon
	eHeaderValuePreStart
	eHeaderValueStart
	eHeaderValue
	eHeaderValueN
	eHeaderEndN
	eHeaderEnd
)

var stateNames = []string{
	"eStart",
	"eVerb",
	"ePathStart",
	"ePath",
	"eVersionStart",
	"eVersionH",
	"eVersionHT",
	"eVersionHTT",
	"eVersionHTTP",
	"eVersionNumberStart",
	"eVersionNumberAll",
	"eVersionNumberAllN",
	"eHeaderStart",
	"eHeaderColon",
	"eHeaderValuePreStart",
	"eHeaderValueStart",
	"eHeaderValue",
	"eHeaderValueN",
	"eHeaderEndN",
	"eHeaderEnd",
}

var ErrBadProto = errors.New("bad protocol")

func (hp *HTTPParser) Parse3(input []byte) error {
	var start int
	var state int
	var numHeader int

	var headerName []byte

	for idx, c := range input {
		// fmt.Printf("> %c - %s\n", rune(c), stateNames[state])
		switch state {
		case eStart:
			start = idx
			state = eVerb
		case eVerb:
			if c == ' ' {
				hp.Method = input[start:idx]
				state = ePathStart
			}
		case ePathStart:
			if c != ' ' {
				start = idx
				state = ePath
			}
		case ePath:
			if c == ' ' {
				hp.Path = input[start:idx]
				state = eVersionH
			}
		case eVersionH:
			if c == 'H' {
				start = idx
				state = eVersionHT
			} else {
				return ErrBadProto
			}
		case eVersionHT:
			if c == 'T' {
				state = eVersionHTT
			} else {
				return ErrBadProto
			}
		case eVersionHTT:
			if c == 'T' {
				state = eVersionHTTP
			} else {
				return ErrBadProto
			}
		case eVersionHTTP:
			if c == 'P' {
				state = eVersionNumberStart
			} else {
				return ErrBadProto
			}
		case eVersionNumberStart:
			state = eVersionNumberAll
		case eVersionNumberAll:
			switch c {
			case '\r':
				hp.Version = input[start:idx]
				state = eVersionNumberAllN
			case '\n':
				hp.Version = input[start:idx]
				state = eHeaderStart
			}
		case eVersionNumberAllN:
			if c == '\n' {
				state = eHeaderStart
			} else {
				return ErrBadProto
			}
		case eHeaderStart:
			switch c {
			case '\r':
				state = eHeaderEndN
			case '\n':
				state = eHeaderEnd
			default:
				start = idx
				state = eHeaderColon
			}
		case eHeaderColon:
			if c == ':' {
				headerName = input[start:idx]
				state = eHeaderValuePreStart
			}
		case eHeaderValuePreStart:
			if c != ' ' {
				return ErrBadProto
			}

			state = eHeaderValueStart
		case eHeaderValueStart:
			start = idx
			state = eHeaderValue
		case eHeaderValue:
			switch c {
			case '\r':
				hp.headers[numHeader] = header{headerName, input[start:idx]}
				numHeader++
				state = eHeaderValueN
			case '\n':
				hp.headers[numHeader] = header{headerName, input[start:idx]}
				numHeader++
				state = eHeaderStart
			}
		case eHeaderValueN:
			if c != '\n' {
				return ErrBadProto
			}

			state = eHeaderStart
		case eHeaderEndN:
			if c != '\n' {
				return ErrBadProto
			}
			return nil
		}
	}

	return ErrBadProto
}

var ErrMissingData = errors.New("missing data")
var ErrUnsupported = errors.New("unsupported http feature")

func (hp *HTTPParser) Parse(input []byte) (err error) {
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
		return ErrMissingData
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
		return ErrMissingData
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
				return errors.Context(ErrBadProto, "missing newline in version")
			}
			headers = i + 1
			ok = true
			break loop
		}
	}

	if !ok {
		return ErrMissingData
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
				return nil
			case ' ', '\t':
				state = 7
			default:
				start = i
				state = 0
			}
		case 6:
			if input[i] != '\n' {
				return ErrBadProto
			}

			return nil
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
		case 3:
			if input[i] != '\n' {
				return ErrBadProto
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

	return ErrMissingData

	/*
		for input[headers] != '\r' {
			for i := headers; i < len(input); i++ {
				if input[i] == ':' {
					for j := i + 2; j < len(input); j++ {
						if input[j] == '\r' {
							hp.headers[h] = header{input[headers:i], input[i+2 : j]}
							h++
							headers = j + 2
							break
						}
					}
					break
				}
			}
		}
	*/

	return nil
}

func (hp *HTTPParser) FindHeader(name []byte) []byte {
	for _, header := range hp.headers {
		if bytes.Equal(header.Name, name) {
			return header.Value
		}
	}

	return nil
}
