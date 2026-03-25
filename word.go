package contestio

func parseWord[T ~string](token []byte) (T, error) { return T(token), nil }

// ScanWord считывает одно или несколько слов (последовательность непробельных байт) из br и сохраняет их по указателям a.
// Возвращает количество считанных слов и ошибку.
//
// Ограничение: Для всех функций ScanWord* длина слова не может превышать размер внутреннего буфера (по умолчанию 4КБ) минус один символ.
// При превышении возвращается ErrTokenTooLong.
func ScanWord[T ~string](br *Reader, a ...*T) (int, error) { return scanVars(br, parseWord, a...) }

// ScanWordLn считывает одно или несколько слов из текущей строки и сохраняет их по указателям a.
// Пропускает оставшуюся часть строки до конца. Возвращает количество считанных слов и ошибку.
func ScanWordLn[T ~string](br *Reader, a ...*T) (int, error) { return scanVarsLn(br, parseWord, a...) }

// ScanWords считывает последовательность слов из br в слайс a.
// Возвращает количество успешно считанных элементов и первую ошибку.
func ScanWords[T ~string](br *Reader, a []T) (int, error) { return scanSlice(br, parseWord, a) }

// ScanWordsLn считывает слова из текущей строки до её конца и добавляет их в слайс a.
// Возвращает итоговый слайс и ошибку (может быть io.EOF).
func ScanWordsLn[S ~[]T, T ~string](br *Reader, a S) (S, error) { return scanSliceLn(br, parseWord, a) }

func printWordsCommon[T ~string](bw *Writer, op writeOpts, a []T) (int, error) {
	_, _ = bw.WriteString(op.Begin)
	for i, v := range a {
		if i > 0 {
			_, _ = bw.WriteString(op.Sep)
		}
		if _, err := bw.WriteString(string(v)); err != nil {
			return i, err
		}
	}
	_, err := bw.WriteString(op.End)
	return len(a), err
}

// PrintWord выводит одну или несколько строк a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintWord[T ~string](bw *Writer, op WO, a ...T) (int, error) {
	return must(printWordsCommon(bw, op, a))
}

// PrintWordLn выводит одну или несколько строк a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintWordLn[T ~string](bw *Writer, a ...T) (int, error) {
	return must(printWordsCommon(bw, lineWO, a))
}

// PrintWords выводит слайс строк a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintWords[T ~string](bw *Writer, op WO, a []T) (int, error) {
	return must(printWordsCommon(bw, op, a))
}

// PrintWordsLn выводит слайс строк a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintWordsLn[T ~string](bw *Writer, a []T) (int, error) {
	return must(printWordsCommon(bw, lineWO, a))
}
