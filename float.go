package contestio

import (
	"strconv"
	"unsafe"
)

// Float обобщает типы чисел с плавающей точкой
type Float interface{ ~float32 | ~float64 }

func _parseFloat[T Float](token []byte) (T, error) {
	bits := int(unsafe.Sizeof(T(0))) << 3
	v, err := strconv.ParseFloat(_unsafeString(token), bits)
	return T(v), err
}
func _parseFloatToPtr[T Float](token []byte, p *T) error { return _parseToPtr(token, _parseFloat, p) }

func _appendFloat[T Float](buf []byte, v T) []byte {
	bitSize := int(unsafe.Sizeof(T(0))) << 3
	return strconv.AppendFloat(buf, float64(v), 'g', -1, bitSize)
}
func _printFloat[T Float](bw *Writer, v T) error { return _writeAppendFunc(bw, _appendFloat[T], v) }

var _ _parseFunc[float64] = _parseFloat[float64]
var _ _parseToFunc[*float64] = _parseFloatToPtr[float64]
var _ _appendValFunc[float64] = _appendFloat[float64]
var _ _printValFunc[float32] = _printFloat[float32]

// ScanFloat считывает одно или несколько чисел с плавающей точкой из br и сохраняет их по указателям a.
// Возвращает количество считанных чисел и ошибку.
func ScanFloat[T Float](br *Reader, a ...*T) (int, error) {
	return _scanVars(br, _parseFloatToPtr, a...)
}

// ScanFloatLn считывает одно или несколько чисел с плавающей точкой из текущей строки и сохраняет их
// по указателям a. Пропускает оставшуюся часть строки до конца. Возвращает количество считанных чисел и ошибку.
func ScanFloatLn[T Float](br *Reader, a ...*T) (int, error) {
	return _scanVarsLn(br, _parseFloatToPtr, a...)
}

// ScanFloats считывает последовательность чисел с плавающей точкой из br в слайс a.
// Возвращает количество успешно считанных элементов и ошибку.
func ScanFloats[T Float](br *Reader, a []T) (int, error) { return _scanSlice(br, _parseFloat, a) }

// ScanFloatsLn считывает числа с плавающей точкой из текущей строки до её конца и добавляет их в слайс a.
// Возвращает итоговый слайс и ошибку.
func ScanFloatsLn[S ~[]T, T Float](br *Reader, a S) (S, error) {
	return _scanSliceLn(br, _parseFloat, a)
}

// PrintFloat выводит одно или несколько чисел с плавающей точкой a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintFloat[T Float](bw *Writer, op WO, a ...T) (int, error) {
	return _printSliceAppend(bw, op, _appendFloat, a)
}

// PrintFloatLn выводит одно или несколько чисел с плавающей точкой a в bw,
// разделяя пробелами и завершая переводом строки. Возвращает количество выведенных элементов и ошибку.
func PrintFloatLn[T Float](bw *Writer, a ...T) (int, error) {
	return _printSliceAppend(bw, _lineWO, _appendFloat, a)
}

// PrintFloats выводит слайс чисел с плавающей точкой a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintFloats[T Float](bw *Writer, op WO, a []T) (int, error) {
	return _printSliceAppend(bw, op, _appendFloat, a)
}

// PrintFloatsLn выводит слайс чисел с плавающей точкой a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintFloatsLn[T Float](bw *Writer, a []T) (int, error) {
	return _printSliceAppend(bw, _lineWO, _appendFloat, a)
}
