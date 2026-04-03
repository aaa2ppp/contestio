package contestio

func _parseWord[T ~string](token []byte) (T, error)       { return T(token), nil }
func _parseWordToPtr[T ~string](token []byte, p *T) error { return _parseToPtr(token, _parseWord, p) }

func _printString[T ~string](bw *Writer, v T) error {
	if _, err := bw.WriteString(string(v)); err != nil {
		return err
	}
	return nil
}

var _ _parseFunc[string] = _parseWord[string]
var _ _parseToFunc[*string] = _parseWordToPtr[string]
var _ _printValFunc[string] = _printString[string]

// ScanWord считывает одно или несколько слов (последовательность непробельных байт) из br и сохраняет их по указателям a.
// Возвращает количество считанных слов и ошибку.
//
// Ограничение: Для всех функций ScanWord* длина слова не может превышать размер внутреннего буфера (по умолчанию 4КБ) минус один символ.
// При превышении возвращается ErrTokenTooLong.
func ScanWord[T ~string](br *Reader, a ...*T) (int, error) {
	return _scanVars(br, _parseWordToPtr, a...)
}

// ScanWordLn считывает одно или несколько слов из текущей строки и сохраняет их по указателям a.
// Пропускает оставшуюся часть строки до конца. Возвращает количество считанных слов и ошибку.
func ScanWordLn[T ~string](br *Reader, a ...*T) (int, error) {
	return _scanVarsLn(br, _parseWordToPtr, a...)
}

// ScanWords считывает последовательность слов из br в слайс a.
// Возвращает количество успешно считанных элементов и первую ошибку.
func ScanWords[T ~string](br *Reader, a []T) (int, error) { return _scanSlice(br, _parseWord, a) }

// ScanWordsLn считывает слова из текущей строки до её конца и добавляет их в слайс a.
// Возвращает итоговый слайс и ошибку (может быть io.EOF).
func ScanWordsLn[S ~[]T, T ~string](br *Reader, a S) (S, error) {
	return _scanSliceLn(br, _parseWord, a)
}

// PrintWord выводит одну или несколько строк a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintWord[T ~string](bw *Writer, op WO, a ...T) (int, error) {
	return _printSlice(bw, op, _printString, a)
}

// PrintWordLn выводит одну или несколько строк a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintWordLn[T ~string](bw *Writer, a ...T) (int, error) {
	return _printSlice(bw, _lineWO, _printString, a)
}

// PrintWords выводит слайс строк a в bw с заданными опциями форматирования.
// Возвращает количество выведенных элементов и ошибку.
func PrintWords[T ~string](bw *Writer, op WO, a []T) (int, error) {
	return _printSlice(bw, op, _printString, a)
}

// PrintWordsLn выводит слайс строк a в bw, разделяя пробелами и завершая переводом строки.
// Возвращает количество выведенных элементов и ошибку.
func PrintWordsLn[T ~string](bw *Writer, a []T) (int, error) {
	return _printSlice(bw, _lineWO, _printString, a)
}
