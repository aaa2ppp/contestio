//go:build parse_int_fast

package contestio

func parseInt[T Int](token []byte) (T, error) { return parseIntFast[T](token) }
