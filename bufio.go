package wildcat

import (
	"bytes"
	"io"
)

type sizedBodyReader struct {
	size int
	rest []byte
	c    io.Reader
}

func (br *sizedBodyReader) Read(buf []byte) (int, error) {
	if br.size == 0 {
		return 0, io.EOF
	}

	if br.rest != nil {
		if len(buf) < len(br.rest) {
			copy(buf, br.rest[:len(buf)])

			br.rest = br.rest[len(buf):]
			br.size -= len(buf)
			return len(buf), nil
		} else {
			l := len(br.rest)
			copy(buf, br.rest)

			br.rest = nil
			br.size -= l
			return l, nil
		}
	}

	n, err := br.c.Read(buf[:br.size])
	if err != nil {
		return 0, err
	}

	br.size -= n
	return n, nil
}

type unsizedBodyReader struct {
	rest []byte
	c    io.Reader
}

func (br *unsizedBodyReader) Read(buf []byte) (int, error) {
	if br.rest != nil {
		if len(buf) < len(br.rest) {
			copy(buf, br.rest[:len(buf)])

			br.rest = br.rest[len(buf):]
			return len(buf), nil
		} else {
			l := len(br.rest)
			copy(buf, br.rest)

			br.rest = nil
			return l, nil
		}
	}

	return br.c.Read(buf)
}

func BodyReader(size int, rest []byte, c io.Reader) io.Reader {
	switch size {
	case 0:
		return bytes.NewReader([]byte(""))
	case -1:
		return &unsizedBodyReader{rest, c}
	default:
		return &sizedBodyReader{size, rest, c}
	}
}
