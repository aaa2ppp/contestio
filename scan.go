package contestio

import (
	"errors"
	"io"
)

type parseFunc[T any] func([]byte) (T, error)
type parseToFunc[T any] func([]byte, T) error

func parseToPtr[T any](token []byte, parse parseFunc[T], p *T) error {
	v, err := parse(token)
	if err != nil {
		return err
	}
	*p = v
	return nil
}

func scanSliceCommon[T any](br *Reader, parse parseFunc[T], a []T) (int, error) {
	for i := range a {
		err := skipSpace(br, false) // don't stop at end of line
		if err != nil {
			if err == io.EOF {
				if i == 0 { // return EOF only if no tokens were read
					return 0, io.EOF
				}
				return i, io.ErrUnexpectedEOF // not all requested data was read
			}
			return i, err
		}
		// success 'skipSpace' ensures that there is at least one non-white character
		token, err := nextToken(br) // always not empty
		if err != nil && err != io.EOF {
			return i, err
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
	n := 0
	for {
		err := skipSpace(br, true) // stop at end of line
		if err != nil {
			if err == EOL {
				return a, nil
			}
			if err == io.EOF {
				if n == 0 { // return EOF only if no tokens were read
					return a, io.EOF
				}
				return a, nil // otherwise interpret as end of line
			}
			return a, err
		}
		// success 'skipSpace' ensures that there is at least one non-white character
		token, err := nextToken(br) // always not empty
		if err != nil && err != io.EOF {
			return a, err
		}
		v, err := parse(token)
		if err != nil {
			return a, err
		}
		a = append(a, v)
		n++
	}
}

func scanVarsCommon[T any](br *Reader, stopAtEol bool, parseTo parseToFunc[T], a []T) (int, error) {
	for i := range a {
		err := skipSpace(br, stopAtEol)
		if err != nil {
			if err == io.EOF {
				if i == 0 { // return EOF only if no tokens were read
					return 0, io.EOF
				}
				return i, io.ErrUnexpectedEOF // not all requested data was read
			}
			return i, err
		}
		// success 'skipSpace' ensures that there is at least one non-white character
		token, err := nextToken(br) // always not empty
		if err != nil && err != io.EOF {
			return i, err
		}
		if err := parseTo(token, a[i]); err != nil {
			return i, err
		}
	}
	return len(a), nil
}

// ErrExpectedEOL возвращается, если не был прочитан ожидаемый конец строки
var ErrExpectedEOL = errors.New("expected end of line")

func scanVarsLnCommon[T any](br *Reader, parseTo parseToFunc[T], a []T) (int, error) {
	n, err := scanVarsCommon(br, true, parseTo, a) // scan to end of line
	if err != nil {
		return n, err
	}
	err = skipSpace(br, true) // stop at end of line
	if err != nil {
		if err == EOL || err == io.EOF { // interpret EOF as end of line
			return n, nil
		}
		return n, err
	}
	return n, ErrExpectedEOL
}

func scanSlice[T any](br *Reader, parse parseFunc[T], a []T) (int, error) {
	return must(scanSliceCommon(br, parse, a))
}

func scanSliceLn[T any](br *Reader, parse parseFunc[T], a []T) ([]T, error) {
	return must(scanSliceLnCommon(br, parse, a))
}

func scanVars[T any](br *Reader, parseTo parseToFunc[T], a ...T) (int, error) {
	return must(scanVarsCommon(br, false, parseTo, a))
}

func scanVarsLn[T any](br *Reader, parseTo parseToFunc[T], a ...T) (int, error) {
	return must(scanVarsLnCommon(br, parseTo, a))
}
