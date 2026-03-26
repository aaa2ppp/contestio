//go:build sugar

package contestio

import (
	"errors"
	"io"
	"reflect"
)

func parseAnyInt[T Int](token []byte, x any) error { // x must be Int pointer
	p := (*T)(reflect.ValueOf(x).UnsafePointer())
	v, err := parseInt[T](token)
	if err != nil {
		return err
	}
	*p = v
	return nil
}

func parseAnyFloat[T Float](token []byte, x any) error { // x must be Float pointer
	p := (*T)(reflect.ValueOf(x).UnsafePointer())
	v, err := parseFloat[T](token)
	if err != nil {
		return err
	}
	*p = v
	return nil
}

func parseAnyWord[T ~string](token []byte, x any) error { // x must be ~string pointer
	p := (*T)(reflect.ValueOf(x).UnsafePointer())
	v, err := parseWord[T](token)
	if err != nil {
		return err
	}
	*p = v
	return nil
}

type parseAnyFunc func(token []byte, p any) error

var parseAnyTab = []parseAnyFunc{
	reflect.Int:     parseAnyInt[int],
	reflect.Int8:    parseAnyInt[int8],
	reflect.Int16:   parseAnyInt[int16],
	reflect.Int32:   parseAnyInt[int32],
	reflect.Int64:   parseAnyInt[int64],
	reflect.Uint:    parseAnyInt[uint],
	reflect.Uint8:   parseAnyInt[uint8],
	reflect.Uint16:  parseAnyInt[uint16],
	reflect.Uint32:  parseAnyInt[uint32],
	reflect.Uint64:  parseAnyInt[uint64],
	reflect.Uintptr: parseAnyInt[uintptr],
	reflect.Float32: parseAnyFloat[float32],
	reflect.Float64: parseAnyFloat[float64],
	reflect.String:  parseAnyWord[string],
}

func scanAnyCommon(br *Reader, stopAtEol bool, a ...any) (int, error) {
	for i, x := range a {
		// Following the behavior of fmt.Fscan, we do not explicitly check for nil pointers.
		// Passing a nil pointer will cause a panic, which is considered a programmer error.
		// This avoids extra overhead in the hot path.

		var parse parseAnyFunc
		t := reflect.TypeOf(x)
		if t.Kind() != reflect.Pointer {
			return i, errors.New("type not a pointer: " + t.String())
		}
		k := t.Elem().Kind()
		if uint(k) < uint(len(parseAnyTab)) {
			parse = parseAnyTab[k]
		}
		if parse == nil {
			return i, errors.New("unknown kind: " + k.String())
		}

		if err := skipSpace(br, stopAtEol); err != nil {
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

		if err := parse(token, x); err != nil {
			return i, err
		}
	}
	return len(a), nil
}

func scanAnyLnCommon(br *Reader, a ...any) (int, error) {
	n, err := scanAnyCommon(br, true, a...) // scan to end of line
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

func ScanAny(br *Reader, a ...any) (int, error)   { return must(scanAnyCommon(br, false, a...)) }
func ScanAnyLn(br *Reader, a ...any) (int, error) { return must(scanAnyLnCommon(br, a...)) }
