//go:build sugar

package contestio

type Floats[T Float] []T

func (s Floats[T]) Slice() []T                     { return s }
func (s Floats[T]) Parse(b []byte) (T, error)      { return parseFloat[T](b) }
func (s Floats[T]) AppendVal(b []byte, v T) []byte { return appendFloat(b, v) }
