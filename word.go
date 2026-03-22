package contestio

func parseWord[T ~string](token []byte) (T, error) { return T(token), nil }

func ScanWords[T ~string](br *Reader, a []T) (int, error)    { return scanSlice(br, parseWord, a) }
func ScanWordsLn[T ~string](br *Reader, a []T) ([]T, error)  { return scanSliceLn(br, parseWord, a) }
func ScanWord[T ~string](br *Reader, a ...*T) (int, error)   { return scanVars(br, parseWord, a...) }
func ScanWordLn[T ~string](br *Reader, a ...*T) (int, error) { return scanVarsLn(br, parseWord, a...) }

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

func PrintWords[T ~string](bw *Writer, op WO, a []T) (int, error) {
	return must(printWordsCommon(bw, op, a))
}
func PrintWordsLn[T ~string](bw *Writer, a []T) (int, error) {
	return must(printWordsCommon(bw, lineWO, a))
}
func PrintWord[T ~string](bw *Writer, op WO, a ...T) (int, error) {
	return must(printWordsCommon(bw, op, a))
}
func PrintWordLn[T ~string](bw *Writer, a ...T) (int, error) {
	return must(printWordsCommon(bw, lineWO, a))
}
