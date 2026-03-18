//go:build !parse_int_std && !parse_int_fast && !parse_int_swar

package contestio

func parseInt[T Int](token []byte) (T, error) { return parseIntBase[T](token) }
