//go:build any

package contestio

import (
	"errors"
	"reflect"
)

type _parseAnyFunc func(token []byte, p any) error
type _printAnyFunc func(bw *Writer, v any) error

func _parseIntToAny[T Int](token []byte, x any) error { // x must be Int pointer
	v, err := _parseInt[T](token)
	if err != nil {
		return err
	}
	p := _getAnyPointer[T](x)
	*p = v
	return nil
}

func _parseFloatToAny[T Float](token []byte, x any) error { // x must be Float pointer
	v, err := _parseFloat[T](token)
	if err != nil {
		return err
	}
	p := _getAnyPointer[T](x)
	*p = v
	return nil
}

func _parseWordToAny[T ~string](token []byte, x any) error { // x must be ~string pointer
	v, err := _parseWord[T](token)
	if err != nil {
		return err
	}
	p := _getAnyPointer[T](x)
	*p = v
	return nil
}

var _parseAnyTab = []_parseAnyFunc{
	reflect.Int:     _parseIntToAny[int],
	reflect.Int8:    _parseIntToAny[int8],
	reflect.Int16:   _parseIntToAny[int16],
	reflect.Int32:   _parseIntToAny[int32],
	reflect.Int64:   _parseIntToAny[int64],
	reflect.Uint:    _parseIntToAny[uint],
	reflect.Uint8:   _parseIntToAny[uint8],
	reflect.Uint16:  _parseIntToAny[uint16],
	reflect.Uint32:  _parseIntToAny[uint32],
	reflect.Uint64:  _parseIntToAny[uint64],
	reflect.Uintptr: _parseIntToAny[uintptr],
	reflect.Float32: _parseFloatToAny[float32],
	reflect.Float64: _parseFloatToAny[float64],
	reflect.String:  _parseWordToAny[string],
}

func _parseToAny(token []byte, x any) error {
	t := reflect.TypeOf(x)
	if t.Kind() != reflect.Pointer {
		return errors.New("type not a pointer: " + t.String())
	}
	k := t.Elem().Kind()
	if uint(k) >= uint(len(_parseAnyTab)) {
		return errors.New("unsupported kind: " + k.String())
	}
	parse := _parseAnyTab[k]
	if parse == nil {
		return errors.New("unsupported kind: " + k.String())
	}
	return parse(token, x)
}

func _printAny(bw *Writer, x any) error {
	t := reflect.TypeOf(x)
	k := t.Kind()
	if k == reflect.Pointer {
		k = t.Elem().Kind()
	}
	if uint(k) >= uint(len(_printAnyTab)) {
		return errors.New("unsupported kind: " + k.String())
	}
	printVal := _printAnyTab[k] // see: any_reflect.go and any_unsafe.go
	if printVal == nil {
		return errors.New("unsupported kind: " + k.String())
	}
	return printVal(bw, x)
}

// ScanAny считывает одно или несколько значений из br и сохраняет их по указателям a.
// Возвращает количество успешно считанных значений и ошибку.
//
// Поддерживаемые типы значений:
//   - любые целые
//   - любые числа с плавающей точкой
//   - любые строки (читаются как слова - последовательность непробельных символов)
//
// Аргументы должны быть указателями на поддерживаемые типы. Передача nil-указателя приводит к панике.
// Функция ведёт себя аналогично fmt.Fscan: пропускает пробелы, читает токены до пробельных символов.
func ScanAny(br *Reader, a ...any) (int, error) { return _scanVars(br, _parseToAny, a...) }

// ScanAnyLn считывает одно или несколько значений из текущей строки и сохраняет их по указателям a.
// Поддерживаемые типы те же, что и в ScanAny. После чтения всех значений пропускает оставшуюся
// часть строки (до символа '\n'). Если после требуемых значений остались другие токены в строке,
// возвращает ErrExpectedEOL. В остальном поведение аналогично ScanAny.
func ScanAnyLn(br *Reader, a ...any) (int, error) { return _scanVarsLn(br, _parseToAny, a...) }

// PrintAny выводит одно или несколько значений a в bw с заданными опциями форматирования.
// Возвращает количество выведенных значений и ошибку.
//
// Поддерживаемые типы значений:
//   - любые целые
//   - любые числа с плавающей точкой
//   - любые строки
//   - и указатели на эти типы (значения разыменовываются)
//
// Передача nil-указателя приводит к панике. Передача неподдерживаемого типа возвращает ошибку.
//
// ВАЖНО: Чтобы избежать лишних выделений памяти (аллокаций):
//  1. Всегда передавайте указатели, а не сами значения.
//  2. Если выводите значения в цикле, объявляйте переменные ДО цикла и переиспользуйте их.
//
// Правильно:
//
//	var x int
//	for i := 0; i < N; i++ {
//		x = data[i]
//		PrintAny(bw, op, &x)
//	}
//
// Неправильно (приводит к аллокациям на каждой итерации):
//
//	for i := 0; i < N; i++ {
//		x := data[i]
//		PrintAny(bw, op, &x)
//	}
//
// Также неправильно (передача значения, а не указателя):
//
//	PrintAny(bw, op, data[i])   // аллокация для каждого значения
func PrintAny(bw *Writer, op WO, a ...any) (int, error) { return _printSlice(bw, op, _printAny, a) }

// PrintAnyLn выводит одно или несколько значений a в bw, разделяя пробелами и завершая переводом строки.
// Работает аналогично PrintAny.
func PrintAnyLn(bw *Writer, a ...any) (int, error) { return _printSlice(bw, _lineWO, _printAny, a) }
