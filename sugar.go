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

type Printer[S ~[]T, T any] interface {
	Slice[S, T]
	Print(bw *Writer, v T) error
}

func PrintSlice[S ~[]T, T any](bw *Writer, op WO, a Printer[S, T]) (int, error) {
	return printSlice(bw, op, a.Print, a.Slice())
}

func PrintSliceLn[S ~[]T, T any](bw *Writer, a Printer[S, T]) (int, error) {
	return printSlice(bw, lineWO, a.Print, a.Slice())
}
