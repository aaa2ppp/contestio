package contestio

func parseWord[T ~string](token []byte) (T, error)       { return T(token), nil }
func parseWordToPtr[T ~string](token []byte, p *T) error { return parseToPtr(token, parseWord, p) }

func printString[T ~string](bw *Writer, v T) error {
	if _, err := bw.WriteString(string(v)); err != nil {
		return err
	}
	return nil
}

var _ parseFunc[string] = parseWord[string]
var _ parseToFunc[*string] = parseWordToPtr[string]
var _ printValFunc[string] = printString[string]

// ScanWord считывает одно или несколько слов (последовательность непробельных байт) из br и сохраняет их по указателям a.
// Возвращает количество считанных слов и ошибку.
//
// Ограничение: Для всех функций ScanWord* длина слова не может превышать размер внутреннего буфера (по умолчанию 4КБ) минус один символ.
// При превышении возвращается ErrTokenTooLong.
func ScanWord[T ~string](br *Reader, a ...*T) (int, error) { return scanVars(br, parseWordToPtr, a...) }

// ScanWordLn считывает одно или несколько слов из текущей строки и сохраняет их по указателям a.
// Пропускает оставшуюся часть строки до конца. Возвращает количество считанных слов и ошибку.
func ScanWordLn[T ~string](br *Reader, a ...*T) (int, error) {
	return scanVarsLn(br, parseWordToPtr, a...)
}

// ScanWords считывает последовательность слов из br в слайс a.
// Возвращает количество успешно считанных элементов и первую ошибку.
func ScanWords[T ~string](br *Reader, a []T) (int, error) { return scanSlice(br, parseWord, a) }

// ScanWordsLn считывает слова из текущей строки до её конца и добавляет их в слайс a.
// Возвращает итоговый слайс и ошибку (может быть io.EOF).
func ScanWordsLn[S ~[]T, T ~string](br *Reader, a S) (S, error) { return scanSliceLn(br, parseWord, a) }

// PrintWord выводит одну или несколько строк a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintWord[T ~string](bw *Writer, op WO, a ...T) (int, error) {
	return printSlice(bw, op, printString, a)
}

// PrintWordLn выводит одну или несколько строк a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintWordLn[T ~string](bw *Writer, a ...T) (int, error) {
	return printSlice(bw, lineWO, printString, a)
}

// PrintWords выводит слайс строк a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintWords[T ~string](bw *Writer, op WO, a []T) (int, error) {
	return printSlice(bw, op, printString, a)
}

// PrintWordsLn выводит слайс строк a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintWordsLn[T ~string](bw *Writer, a []T) (int, error) {
	return printSlice(bw, lineWO, printString, a)
}
