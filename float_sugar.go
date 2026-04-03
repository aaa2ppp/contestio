//go:build sugar

package contestio

type Floats[T Float] []T

func (s Floats[T]) Slice() Floats[T]            { return s }
func (s Floats[T]) Parse(b []byte) (T, error)   { return _parseFloat[T](b) }
func (s Floats[T]) Print(bw *Writer, v T) error { return _printFloat(bw, v) }

var _ Parser[Floats[float64], float64] = Floats[float64]{}
var _ Printer[Floats[float64], float64] = Floats[float64]{}
