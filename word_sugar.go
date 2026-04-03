//go:build sugar

package contestio

type Words[T ~string] []T

func (s Words[T]) Slice() Words[T]                { return s }
func (s Words[T]) Parse(b []byte) (T, error)      { return _parseWord[T](b) }
func (s Words[T]) AppendVal(b []byte, v T) []byte { return append(b, v...) }
