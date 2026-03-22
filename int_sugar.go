//go:build sugar

package contestio

type Ints[T Int] []T

func (s Ints[T]) Slice() Ints[T]                 { return s }
func (s Ints[T]) Parse(b []byte) (T, error)      { return parseInt[T](b) }
func (s Ints[T]) AppendVal(b []byte, v T) []byte { return appendInt(b, v) }
