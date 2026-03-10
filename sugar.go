//go:build sugar

package contestio

// Обобщение ScanSlice и PrintSlice слайсов для любого типа.
// Вместо специфических функции Scan[Type]s/Scan[Type]sLn можно использовать ScanSlice/ScanSliceLn.
// Возможно кому-то будет это удобно. Мне кажется избыточным "сахором".
//
// NOTE: При использовании этих функций, будет одна аллокация 16 байт на создание интерфейса.
// И небольшая просадка по производительности из-за использования виртуальных методов.

type Slice[T any] interface {
	Slice() []T
}

type Parser[T any] interface {
	Slice[T]
	Parse([]byte) (T, error)
}

func ScanSlice[T any](br *Reader, a Parser[T]) (int, error) {
	return scanSlice(br, a.Parse, a.Slice())
}

func ScanSliceLn[T any](br *Reader, a Parser[T]) ([]T, error) {
	return scanSliceLn(br, a.Parse, a.Slice())
}

type ValAppender[T any] interface {
	Slice[T]
	AppendVal(b []byte, v T) []byte
}

func PrintSlice[T any](bw *Writer, op WO, a ValAppender[T]) (int, error) {
	return printSlice(bw, op, a.AppendVal, a.Slice())
}

func PrintSliceLn[T any](bw *Writer, a ValAppender[T]) (int, error) {
	return printSliceLn(bw, a.AppendVal, a.Slice())
}
