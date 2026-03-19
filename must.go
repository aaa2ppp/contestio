//go:build must

package contestio

import "io"

// must panic on error, except for io.EOF
func must[T any](v T, err error) (T, error) {
	if err != nil && err != io.EOF {
		panic(err)
	}
	return v, err
}
