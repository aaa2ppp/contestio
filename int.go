package contestio

import (
	"math"
	"strconv"
	"unsafe"
)

type (
	Sign interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64
	}
	Unsig interface {
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
	}
	Int interface{ Sign | Unsig }
)

func parseIntStd[T Int](b []byte) (T, error) {
	signed := ^T(0) < 0
	bitSize := int(unsafe.Sizeof(T(0))) << 3
	if signed {
		v, err := strconv.ParseInt(unsafeString(b), 10, bitSize)
		return T(v), err
	}
	v, err := strconv.ParseUint(unsafeString(b), 10, bitSize)
	return T(v), err
}

// A NumError records a failed conversion of Int
type IntError struct {
	Num string
	Err error
}

func (e *IntError) Error() string { return "parseInt: " + strconv.Quote(e.Num) + ": " + e.Err.Error() }
func (e *IntError) Unwrap() error { return e.Err }

func parseIntBase[T Int](token []byte) (T, error) {
	var unsigned = ^T(0) >= 0 // true for unsigned T

	orig := token
	if len(orig) == 0 {
		return 0, &IntError{string(orig), strconv.ErrSyntax}
	}

	// handle sign
	if orig[0] == '-' || orig[0] == '+' {
		if unsigned {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		token = token[1:]
		if len(token) == 0 {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
	}

	// parse to uint64 (no fast-path: len<20; see parse_int_fast.go)
	var u64 uint64
	for _, digit := range token {
		digit -= '0'
		if digit > 9 {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		if u64 < math.MaxUint64/10 || (u64 == math.MaxUint64/10 && digit <= math.MaxUint64%10) {
			u64 = u64*10 + uint64(digit)
			continue
		}
		return 0, &IntError{string(orig), strconv.ErrRange}
	}

	if unsigned {
		if u64 > uint64(^T(0)) {
			return 0, &IntError{string(orig), strconv.ErrRange}
		}
		return T(u64), nil
	}

	// signed range check
	bits := int(unsafe.Sizeof(T(0))) << 3
	absMin := uint64(1) << (bits - 1) // |min(T)|
	if orig[0] == '-' {
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

func parseIntFast[T Int](token []byte) (T, error) {
	var unsigned = ^T(0) >= 0 // true для unsigned T

	orig := token
	if len(orig) == 0 {
		return 0, &IntError{string(orig), strconv.ErrSyntax}
	}

	// handle sign
	if orig[0] == '-' || orig[0] == '+' {
		if unsigned {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		token = token[1:]
		if len(token) == 0 {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
	}

	// trim leading zeros
	for len(token) > 0 && token[0] == '0' {
		token = token[1:]
	}
	if len(token) == 0 {
		return 0, nil // was "0...0"
	}

	// parse to uint64: учитываем, что 10^19 < MaxUint64 < 10^20
	var u64 uint64
	switch {
	case len(token) < 20:
		// переполнение невозможно
		for _, digit := range token {
			digit -= '0'
			if digit > 9 {
				return 0, &IntError{string(orig), strconv.ErrSyntax}
			}
			u64 = u64*10 + uint64(digit)
		}
	case len(token) == 20:
		// проверяем только последний разряд
		for _, digit := range token[:19] {
			digit -= '0'
			if digit > 9 {
				return 0, &IntError{string(orig), strconv.ErrSyntax}
			}
			u64 = u64*10 + uint64(digit)
		}
		digit := token[19] - '0'
		if digit > 9 {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		if u64 < math.MaxUint64/10 || (u64 == math.MaxUint64/10 && digit <= math.MaxUint64%10) {
			u64 = u64*10 + uint64(digit)
		} else {
			return 0, &IntError{string(orig), strconv.ErrRange}
		}
	default: // len(token) > 20
		// гарантированное переполнение
		return 0, &IntError{string(orig), strconv.ErrRange}
	}

	if unsigned {
		if u64 > uint64(^T(0)) {
			return 0, &IntError{string(orig), strconv.ErrRange}
		}
		return T(u64), nil
	}

	// signed range check
	bits := int(unsafe.Sizeof(T(0))) << 3
	absMin := uint64(1) << (bits - 1) // |min(T)|
	if orig[0] == '-' {
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

func appendInt[T Int](buf []byte, v T) []byte {
	signed := ^T(0) < 0
	if signed {
		return strconv.AppendInt(buf, int64(v), 10)
	} else {
		return strconv.AppendUint(buf, uint64(v), 10)
	}
}

type Ints[T Int] []T

func (s Ints[T]) Slice() []T                     { return s }
func (s Ints[T]) Parse(b []byte) (T, error)      { return parseInt[T](b) }
func (s Ints[T]) AppendVal(b []byte, v T) []byte { return appendInt(b, v) }

// ScanInts считывает последовательность целых чисел из br в слайс a.
// Возвращает количество успешно считанных элементов и первую ошибку.
func ScanInts[T Int](br *Reader, a []T) (int, error) { return scanSlice(br, parseInt, a) }

// ScanIntsLn считывает целые числа из текущей строки до её конца и добавляет их в слайс a.
// Возвращает итоговый слайс и ошибку (может быть io.EOF).
func ScanIntsLn[T Int](br *Reader, a []T) ([]T, error) { return scanSliceLn(br, parseInt, a) }

// ScanInt считывает одно или несколько целых чисел из br и сохраняет их по указателям a.
// Возвращает количество считанных чисел и ошибку.
func ScanInt[T Int](br *Reader, a ...*T) (int, error) { return scanVars(br, parseInt, a...) }

// ScanIntLn считывает одно или несколько целых чисел из текущей строки и сохраняет их по
// указателям a. Пропускает оставшуюся часть строки до конца. Возвращает количество считанных
// чисел и ошибку.
func ScanIntLn[T Int](br *Reader, a ...*T) (int, error) { return scanVarsLn(br, parseInt, a...) }

// PrintInts выводит слайс целых чисел a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintInts[T Int](bw *Writer, op WO, a []T) (int, error) { return printSlice(bw, op, appendInt, a) }

// PrintIntsLn выводит слайс целых чисел a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintIntsLn[T Int](bw *Writer, a []T) (int, error) { return printSliceLn(bw, appendInt, a) }

// PrintInt выводит одно или несколько целых чисел a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintInt[T Int](bw *Writer, op WO, a ...T) (int, error) {
	return printVals(bw, op, appendInt, a...)
}

// PrintIntLn выводит одно или несколько целых чисел a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintIntLn[T Int](bw *Writer, a ...T) (int, error) { return printValsLn(bw, appendInt, a...) }
