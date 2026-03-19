package contestio

import (
	"errors"
	"io"
)

type parseFunc[T any] func([]byte) (T, error)

func scanSliceCommon[T any](br *Reader, parse parseFunc[T], a []T) (int, error) {
	var eof bool
	for i := range a {
		if eof {
			return i, io.ErrUnexpectedEOF
		}
		if err := skipSpace(br, false); err != nil {
			if err == io.EOF {
				return i, io.ErrUnexpectedEOF
			}
			return i, err
		}
		token, err := nextToken(br)
		if err != nil {
			if err != io.EOF {
				return i, err
			}
			if len(token) == 0 {
				return i, io.EOF
			}
			eof = true
		}
		v, err := parse(token)
		if err != nil {
			return i, err
		}
		a[i] = v
	}
	return len(a), nil
}

func scanSliceLnCommon[T any](br *Reader, parse parseFunc[T], a []T) ([]T, error) {
	var eof bool
	n := 0
	for {
		if eof {
			if n == 0 {
				return a, io.ErrUnexpectedEOF
			}
			return a, nil
		}
		if err := skipSpace(br, true); err != nil {
			if err == EOL {
				return a, nil
			}
			if err == io.EOF {
				if n == 0 {
					return a, io.EOF
				}
				return a, nil
			}
			return a, err
		}
		token, err := nextToken(br)
		if err != nil {
			if err != io.EOF {
				return a, err
			}
			if len(token) == 0 {
				return a, io.EOF
			}
			eof = true
		}
		v, err := parse(token)
		if err != nil {
			return a, err
		}
		a = append(a, v)
		n++
	}
}

func scanVarsCommon[T any](br *Reader, stopAtEol bool, parse parseFunc[T], a ...*T) (int, error) {
	var eof bool
	for i := range a {
		if eof {
			return i, io.ErrUnexpectedEOF
		}
		if err := skipSpace(br, stopAtEol); err != nil {
			if err == io.EOF {
				return i, io.ErrUnexpectedEOF
			}
			return i, err
		}
		token, err := nextToken(br)
		if err != nil {
			if err != io.EOF {
				return i, err
			}
			if len(token) == 0 {
				return i, io.ErrUnexpectedEOF
			}
			eof = true
		}
		v, err := parse(token)
		if err != nil {
			return i, err
		}
		*a[i] = v
	}
	return len(a), nil
}

// ErrExpectedEOL возвращается, если не был прочитан ожидаемый конец строки
var ErrExpectedEOL = errors.New("expected end of line")

func scanVarsLnCommon[T any](br *Reader, parser func([]byte) (T, error), a ...*T) (int, error) {
	n, err := scanVarsCommon(br, true, parser, a...)
	if err != nil {
		return n, err
	}
	err = skipSpace(br, true)
	if err == EOL || err == io.EOF {
		return n, nil
	}
	if err != nil {
		return n, err
	}
	return n, ErrExpectedEOL
}

func scanSlice[T any](br *Reader, parse parseFunc[T], a []T) (int, error) {
	return scanSliceCommon(br, parse, a)
}

func scanSliceLn[T any](br *Reader, parse parseFunc[T], a []T) ([]T, error) {
	return scanSliceLnCommon(br, parse, a)
}

func scanVars[T any](br *Reader, parser func([]byte) (T, error), a ...*T) (int, error) {
	return scanVarsCommon(br, false, parser, a...)
}

func scanVarsLn[T any](br *Reader, parser func([]byte) (T, error), a ...*T) (int, error) {
	return scanVarsLnCommon(br, parser, a...)
}
