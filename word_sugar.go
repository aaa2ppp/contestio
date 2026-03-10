//go:build sugar

package contestio

type Words[T Word] []T

func (s Words[T]) Slice() []T                     { return s }
func (s Words[T]) Parse(b []byte) (T, error)      { return parseWord[T](b) }
func (s Words[T]) AppendVal(b []byte, v T) []byte { return append(b, v...) }
