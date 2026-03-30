//go:build any

package contestio

import (
	"errors"
	"reflect"
)

type parseAnyFunc func(token []byte, p any) error
type printAnyFunc func(bw *Writer, v any) error

func parseAnyInt[T Int](token []byte, x any) error { // x must be Int pointer
	v, err := parseInt[T](token)
	if err != nil {
		return err
	}
	p := getAnyPointer[T](x)
	*p = v
	return nil
}

func parseAnyFloat[T Float](token []byte, x any) error { // x must be Float pointer
	v, err := parseFloat[T](token)
	if err != nil {
		return err
	}
	p := getAnyPointer[T](x)
	*p = v
	return nil
}

func parseAnyWord[T ~string](token []byte, x any) error { // x must be ~string pointer
	v, err := parseWord[T](token)
	if err != nil {
		return err
	}
	p := getAnyPointer[T](x)
	*p = v
	return nil
}

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

func parseAnyTo(token []byte, x any) error {
	t := reflect.TypeOf(x)
	if t.Kind() != reflect.Pointer {
		return errors.New("type not a pointer: " + t.String())
	}
	k := t.Elem().Kind()
	if uint(k) >= uint(len(parseAnyTab)) {
		return errors.New("unsupported kind: " + k.String())
	}
	parse := parseAnyTab[k]
	if parse == nil {
		return errors.New("unsupported kind: " + k.String())
	}
	return parse(token, x)
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
func ScanAny(br *Reader, a ...any) (int, error) { return scanVars(br, parseAnyTo, a...) }

// ScanAnyLn считывает одно или несколько значений из текущей строки и сохраняет их по указателям a.
// Поддерживаемые типы те же, что и в ScanAny. После чтения всех значений пропускает оставшуюся
// часть строки (до символа '\n'). Если после требуемых значений остались другие токены в строке,
// возвращает ErrExpectedEOL. В остальном поведение аналогично ScanAny.
func ScanAnyLn(br *Reader, a ...any) (int, error) { return scanVarsLn(br, parseAnyTo, a...) }

func printAny(bw *Writer, x any) error {
	t := reflect.TypeOf(x)
	k := t.Kind()
	if k == reflect.Pointer {
		k = t.Elem().Kind()
	}
	if uint(k) >= uint(len(printAnyTab)) {
		return errors.New("unsupported kind: " + k.String())
	}
	printVal := printAnyTab[k]
	if printVal == nil {
		return errors.New("unsupported kind: " + k.String())
	}
	return printVal(bw, x)
}

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
func PrintAny(bw *Writer, op WO, a ...any) (int, error) { return printVals(bw, op, printAny, a...) }

// PrintAnyLn выводит одно или несколько значений a в bw, разделяя пробелами и завершая переводом строки.
// Работает аналогично PrintAny.
func PrintAnyLn(bw *Writer, a ...any) (int, error) { return printValsLn(bw, printAny, a...) }
