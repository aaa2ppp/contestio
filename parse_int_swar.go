//go:build parse_int_swar

package contestio

import "unsafe"

func parseInt[T Int](token []byte) (T, error) { return parseIntSwar[T](token) }
