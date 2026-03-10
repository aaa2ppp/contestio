//go:build parse_int_std

package contestio

func parseInt[T Int](token []byte) (T, error) { return parseIntStd[T](token) }
