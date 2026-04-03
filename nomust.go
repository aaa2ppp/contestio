//go:build !must

package contestio

func _must[T any](v T, err error) (T, error) { return v, err } // do nothing
