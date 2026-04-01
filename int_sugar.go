//go:build sugar

package contestio

type Ints[T Int] []T

func (s Ints[T]) Slice() Ints[T]              { return s }
func (s Ints[T]) Parse(b []byte) (T, error)   { return parseInt[T](b) }
func (s Ints[T]) Print(bw *Writer, v T) error { return printInt(bw, v) }

var _ Parser[Ints[int], int] = Ints[int]{}
var _ Printer[Ints[int], int] = Ints[int]{}
