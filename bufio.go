package wildcat

import "io"

type sizedBodyReader struct {
	size int64
	rest []byte
	c    io.ReadCloser
}

func (br *sizedBodyReader) Read(buf []byte) (int, error) {
	if br.size == 0 {
		return 0, io.EOF
	}

	if br.rest != nil {
		if len(buf) < len(br.rest) {
			copy(buf, br.rest[:len(buf)])

			br.rest = br.rest[len(buf):]
			br.size -= int64(len(buf))
			return len(buf), nil
		} else {
			l := len(br.rest)
			copy(buf, br.rest)

			br.rest = nil
			br.size -= int64(l)
			return l, nil
		}
	}

	n, err := br.c.Read(buf[:br.size])
	if err != nil {
		return 0, err
	}

	br.size -= int64(n)
	return n, nil
}

func (br *sizedBodyReader) Close() error {
	return br.c.Close()
}

type unsizedBodyReader struct {
	rest []byte
	c    io.ReadCloser
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

func (br *unsizedBodyReader) Close() error {
	return br.c.Close()
}

func BodyReader(size int64, rest []byte, c io.ReadCloser) io.ReadCloser {
	switch size {
	case 0:
		return nil
	case -1:
		return &unsizedBodyReader{rest, c}
	default:
		return &sizedBodyReader{size, rest, c}
	}
}
