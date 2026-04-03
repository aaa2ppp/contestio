//go:build must

package contestio

import "io"

// _must panic on error, except for io.EOF
func _must[T any](v T, err error) (T, error) {
	if err != nil && err != io.EOF {
		panic(err)
	}
	return v, err
}
