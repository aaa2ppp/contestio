package contestio

import (
	"errors"
)

// ErrTokenTooLong возвращается, если считываемый токен превышает размер внутреннего буфера.
var ErrTokenTooLong = errors.New("token too long")

func nextToken(br *Reader) ([]byte, error) {
	var buf []byte
	var err error
	i := 0
	fast := br.Buffered() > 0

	for i < br.Size() {
		if fast {
			// Fast path without fill
			buf, _ = br.Peek(br.Buffered())
		} else {
			// Slow path with fill
			buf, err = br.Peek(br.Buffered() + 1)
			if err != nil {
				_, _ = br.Discard(len(buf))
				return buf, err
			}
		}
		buf = buf[:br.Buffered()]
		for ; i < len(buf); i++ {
			if isSpace(buf[i]) {
				_, _ = br.Discard(i)
				return buf[:i], nil
			}
		}
		fast = false
	}
	_, _ = br.Discard(len(buf))
	return buf, ErrTokenTooLong
}

var spaceTab = [256]bool{
	' ':  true,
	'\t': true,
	'\r': true,
	'\n': true,
}

func isSpace(c byte) bool { return spaceTab[c] }

var EOL = errors.New("EOL")

func skipSpace(br *Reader, stopAtNewLine bool) error {
	var buf []byte
	var err error
	fast := br.Buffered() > 0

	for {
		if fast {
			// Fast path without fill
			buf, _ = br.Peek(br.Buffered())
		} else {
			// Slow path with fill
			buf, err = br.Peek(br.Buffered() + 1)
			if err != nil {
				return err
			}
		}
		buf = buf[:br.Buffered()]
		for i, c := range buf {
			if stopAtNewLine && c == '\n' {
				_, _ = br.Discard(i + 1)
				return EOL
			}
			if !isSpace(c) {
				_, _ = br.Discard(i)
				return nil
			}
		}
		_, _ = br.Discard(len(buf))
		fast = false
	}
}

func skipToNewLine(br *Reader) error {
	var buf []byte
	var err error
	fast := br.Buffered() > 0

	for {
		if fast {
			// Fast path without fill
			buf, _ = br.Peek(br.Buffered())
		} else {
			// Slow path with fill
			buf, err = br.Peek(br.Buffered() + 1)
			if err != nil {
				return err
			}
		}
		buf = buf[:br.Buffered()]
		for i, c := range buf {
			if c == '\n' {
				_, _ = br.Discard(i + 1)
				return nil
			}
		}
		_, _ = br.Discard(len(buf))
		fast = false
	}
}
