//go:build any

package contestio

import (
	"errors"
	"io"
	"reflect"
)

type parseAnyFunc func(token []byte, p any) error
type appendAnyFunc func(b []byte, v any) []byte

func scanAnyCommon(br *Reader, stopAtEol bool, a ...any) (int, error) {
	for i, x := range a {
		// Following the behavior of fmt.Fscan, we do not explicitly check for nil pointers.
		// Passing a nil pointer will cause a panic, which is considered a programmer error.
		// This avoids extra overhead in the hot path.

		t := reflect.TypeOf(x)
		if t.Kind() != reflect.Pointer {
			return i, errors.New("type not a pointer: " + t.String())
		}
		k := t.Elem().Kind()
		if uint(k) >= uint(len(parseAnyTab)) {
			return i, errors.New("unsupported kind: " + k.String())
		}
		parseAny := parseAnyTab[k]
		if parseAny == nil {
			return i, errors.New("unsupported kind: " + k.String())
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

		if err := parseAny(token, x); err != nil {
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

func printAnyCommon(bw *Writer, op writeOpts, a ...any) (int, error) {
	var buf []byte

	_, _ = bw.WriteString(op.Begin)

	for i, x := range a {
		if i > 0 {
			_, _ = bw.WriteString(op.Sep)
		}

		t := reflect.TypeOf(x)
		k := t.Kind()
		if k == reflect.Pointer {
			k = t.Elem().Kind()
		}

		if k == reflect.String {
			v := getStringValue(x)
			if _, err := bw.WriteString(v); err != nil {
				return i, err
			}
			continue
		}

		if uint(k) >= uint(len(appendAnyTab)) {
			return i, errors.New("unsupported kind: " + k.String())
		}
		appendVal := appendAnyTab[k]
		if appendVal == nil {
			return i, errors.New("unsupported kind: " + k.String())
		}

		if bw.Available() < len(bw.scratch) {
			buf = bw.scratch[:0]
		} else {
			buf = bw.AvailableBuffer()
		}

		buf = appendVal(buf, a[i])
		if _, err := bw.Write(buf); err != nil {
			return i, err
		}
	}

	_, err := bw.WriteString(op.End)
	return len(a), err
}

func PrintAny(bw *Writer, op WO, a ...any) (int, error) { return must(printAnyCommon(bw, op, a...)) }
func PrintAnyLn(bw *Writer, a ...any) (int, error)      { return must(printAnyCommon(bw, lineWO, a...)) }
