package contestio

import (
	"strconv"
	"unsafe"
)

// Int обобщает все целочисленные типы
type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// A NumError records a failed conversion of Int
type IntError struct {
	Num string
	Err error
}

func (e *IntError) Error() string { return "parseInt: " + strconv.Quote(e.Num) + ": " + e.Err.Error() }
func (e *IntError) Unwrap() error { return e.Err }

func parseInt[T Int](token []byte) (T, error) {
	var unsigned = ^T(0) >= 0 // true for unsigned T

	orig := token
	if len(orig) == 0 {
		return 0, &IntError{string(orig), strconv.ErrSyntax}
	}

	// handle sign
	firstChar := orig[0]
	if firstChar == '-' || firstChar == '+' {
		if unsigned {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		token = orig[1:]
		if len(token) == 0 {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
	}

	// parse to uint64
	var u64 uint64
	if len(token) < 20 {
		// fast path without overflows check
		for i := uint(0); i < uint(len(token)); i++ {
			digit := token[i] - '0'
			if digit > 9 {
				return 0, &IntError{string(orig), strconv.ErrSyntax}
			}
			u64 = u64*10 + uint64(digit)
		}
	} else {
		var err error
		u64, err = strconv.ParseUint(unsafeString(token), 10, 64)
		if err != nil {
			if numErr, ok := err.(*strconv.NumError); ok {
				return 0, &IntError{string(orig), numErr.Err}
			}
			return 0, &IntError{string(orig), err}
		}
	}

	// unsigned range check
	if unsigned {
		if u64 > uint64(^T(0)) {
			return 0, &IntError{string(orig), strconv.ErrRange}
		}
		return T(u64), nil
	}

	// signed range check
	bits := int(unsafe.Sizeof(T(0))) << 3
	absMin := uint64(1) << (bits - 1) // |min(T)|
	if firstChar == '-' {
		if u64 > absMin {
			return 0, &IntError{string(orig), strconv.ErrRange}
		}
		return -T(u64), nil // two's complement: корректно при u64==absMin
	}
	if u64 >= absMin {
		return 0, &IntError{string(orig), strconv.ErrRange}
	}
	return T(u64), nil
}
func parseIntToPtr[T Int](token []byte, p *T) error { return parseToPtr(token, parseInt, p) }

func appendInt[T Int](buf []byte, v T) []byte {
	signed := ^T(0) < 0
	if signed {
		return strconv.AppendInt(buf, int64(v), 10)
	} else {
		return strconv.AppendUint(buf, uint64(v), 10)
	}
}
func printInt[T Int](bw *Writer, v T) error { return writeAppendFunc(bw, appendInt[T], v) }

var _ parseFunc[int] = parseInt[int]
var _ parseToFunc[*int] = parseIntToPtr[int]
var _ appendValFunc[int] = appendInt[int]
var _ printValFunc[int] = printInt[int]

// ScanInt считывает одно или несколько целых чисел из br и сохраняет их по указателям a.
// Возвращает количество считанных чисел и ошибку.
func ScanInt[S []*T, T Int](br *Reader, a ...*T) (int, error) {
	return scanVars(br, parseIntToPtr, a...)
}

// ScanIntLn считывает одно или несколько целых чисел из текущей строки и сохраняет их по
// указателям a. Пропускает оставшуюся часть строки до конца. Возвращает количество считанных
// чисел и ошибку.
func ScanIntLn[T Int](br *Reader, a ...*T) (int, error) { return scanVarsLn(br, parseIntToPtr, a...) }

// ScanInts считывает последовательность целых чисел из br в слайс a.
// Возвращает количество успешно считанных элементов и первую ошибку.
func ScanInts[T Int](br *Reader, a []T) (int, error) { return scanSlice(br, parseInt, a) }

// ScanIntsLn считывает целые числа из текущей строки до её конца и добавляет их в слайс a.
// Возвращает итоговый слайс и ошибку (может быть io.EOF).
func ScanIntsLn[S ~[]T, T Int](br *Reader, a S) (S, error) { return scanSliceLn(br, parseInt, a) }

// PrintInt выводит одно или несколько целых чисел a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintInt[T Int](bw *Writer, op WO, a ...T) (int, error) {
	return printVals(bw, op, printInt, a...)
}

// PrintIntLn выводит одно или несколько целых чисел a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintIntLn[T Int](bw *Writer, a ...T) (int, error) { return printValsLn(bw, printInt, a...) }

// PrintInts выводит слайс целых чисел a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintInts[T Int](bw *Writer, op WO, a []T) (int, error) { return printSlice(bw, op, printInt, a) }

// PrintIntsLn выводит слайс целых чисел a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintIntsLn[T Int](bw *Writer, a []T) (int, error) { return printSliceLn(bw, printInt, a) }
