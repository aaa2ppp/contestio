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
func ScanAny(br *Reader, a ...any) (int, error) { return must(scanAnyCommon(br, false, a...)) }

// ScanAnyLn считывает одно или несколько значений из текущей строки и сохраняет их по указателям a.
// Поддерживаемые типы те же, что и в ScanAny. После чтения всех значений пропускает оставшуюся
// часть строки (до символа '\n'). Если после требуемых значений остались другие токены в строке,
// возвращает ErrExpectedEOL. В остальном поведение аналогично ScanAny.
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
			v := getAnyString(x)
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
func PrintAny(bw *Writer, op WO, a ...any) (int, error) { return must(printAnyCommon(bw, op, a...)) }

// PrintAnyLn выводит одно или несколько значений a в bw, разделяя пробелами и завершая переводом строки.
// Работает аналогично PrintAny.
func PrintAnyLn(bw *Writer, a ...any) (int, error) { return must(printAnyCommon(bw, lineWO, a...)) }
