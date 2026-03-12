package contestio

import (
	"strconv"
	"unsafe"
)

// Float обобщает типы чисел с плавающей точкой
type Float interface{ ~float32 | ~float64 }

func parseFloat[T Float](token []byte) (T, error) {
	bits := int(unsafe.Sizeof(T(0))) << 3
	v, err := strconv.ParseFloat(unsafeString(token), bits)
	return T(v), err
}

func appendFloat[T Float](buf []byte, v T) []byte {
	bitSize := int(unsafe.Sizeof(T(0))) << 3
	return strconv.AppendFloat(buf, float64(v), 'g', -1, bitSize)
}

// ScanFloats считывает последовательность чисел с плавающей точкой из br в слайс a.
// Возвращает количество успешно считанных элементов и ошибку.
func ScanFloats[T Float](br *Reader, a []T) (int, error) { return scanSlice(br, parseFloat, a) }

// ScanFloatsLn считывает числа с плавающей точкой из текущей строки до её конца и добавляет их в слайс a.
// Возвращает итоговый слайс и ошибку.
func ScanFloatsLn[T Float](br *Reader, a []T) ([]T, error) { return scanSliceLn(br, parseFloat, a) }

// ScanFloat считывает одно или несколько чисел с плавающей точкой из br и сохраняет их по указателям a.
// Возвращает количество считанных чисел и ошибку.
func ScanFloat[T Float](br *Reader, a ...*T) (int, error) { return scanVars(br, parseFloat, a...) }

// ScanFloatLn считывает одно или несколько чисел с плавающей точкой из текущей строки и сохраняет их
// по указателям a. Пропускает оставшуюся часть строки до конца. Возвращает количество считанных чисел и ошибку.
func ScanFloatLn[T Float](br *Reader, a ...*T) (int, error) { return scanVarsLn(br, parseFloat, a...) }

// PrintFloats выводит слайс чисел с плавающей точкой a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintFloats[T Float](bw *Writer, op WO, a []T) (int, error) {
	return printSlice(bw, op, appendFloat, a)
}

// PrintFloatsLn выводит слайс чисел с плавающей точкой a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintFloatsLn[T Float](bw *Writer, a []T) (int, error) { return printSliceLn(bw, appendFloat, a) }

// PrintFloat выводит одно или несколько чисел с плавающей точкой a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintFloat[T Float](bw *Writer, op WO, a ...T) (int, error) {
	return printVals(bw, op, appendFloat, a...)
}

// PrintFloatLn выводит одно или несколько чисел с плавающей точкой a в bw,
// разделяя пробелами и завершая переводом строки. Возвращает количество выведенных элементов и ошибку.
func PrintFloatLn[T Float](bw *Writer, a ...T) (int, error) {
	return printValsLn(bw, appendFloat, a...)
}
