package contestio

type Word interface{ ~string }

func parseWord[T Word](token []byte) (T, error) { return T(token), nil }
func appendWord[T Word](b []byte, v T) []byte   { return append(b, v...) }

func ScanWords[T Word](br *Reader, a []T) (int, error)    { return scanSlice(br, parseWord, a) }
func ScanWordsLn[T Word](br *Reader, a []T) ([]T, error)  { return scanSliceLn(br, parseWord, a) }
func ScanWord[T Word](br *Reader, a ...*T) (int, error)   { return scanVars(br, parseWord, a...) }
func ScanWordLn[T Word](br *Reader, a ...*T) (int, error) { return scanVarsLn(br, parseWord, a...) }

func PrintWords[T Word](bw *Writer, op WO, a []T) (int, error) {
	return printSlice(bw, op, appendWord, a)
}
func PrintWordsLn[T Word](bw *Writer, a []T) (int, error) { return printSliceLn(bw, appendWord, a) }
func PrintWord[T Word](bw *Writer, op WO, a ...T) (int, error) {
	return printVals(bw, op, appendWord, a...)
}
func PrintWordLn[T Word](bw *Writer, a ...T) (int, error) { return printValsLn(bw, appendWord, a...) }
