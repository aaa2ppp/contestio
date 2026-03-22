//go:build sugar

package contestio

// Обобщение ScanSlice и PrintSlice для слайсов любого типа.
// Вместо специфических функции Scan[Type]s/Scan[Type]sLn можно использовать ScanSlice/ScanSliceLn.
// Возможно кому-то будет это удобно. Мне кажется избыточным "сахором".
//
// NOTE: При использовании этих функций, будет одна аллокация 16 байт на создание интерфейса.
// И небольшая просадка по производительности из-за использования виртуальных методов.

type Slice[S ~[]T, T any] interface {
	Slice() S
}

type Parser[S ~[]T, T any] interface {
	Slice[S, T]
	Parse([]byte) (T, error)
}

func ScanSlice[S ~[]T, T any](br *Reader, a Parser[S, T]) (int, error) {
	return scanSlice(br, a.Parse, a.Slice())
}

func ScanSliceLn[S ~[]T, T any](br *Reader, a Parser[S, T]) (S, error) {
	return scanSliceLn(br, a.Parse, a.Slice())
}

type ValAppender[S ~[]T, T any] interface {
	Slice[S, T]
	AppendVal(b []byte, v T) []byte
}

func PrintSlice[S ~[]T, T any](bw *Writer, op WO, a ValAppender[S, T]) (int, error) {
	return printSlice(bw, op, a.AppendVal, a.Slice())
}

func PrintSliceLn[S ~[]T, T any](bw *Writer, a ValAppender[S, T]) (int, error) {
	return printSliceLn(bw, a.AppendVal, a.Slice())
}
