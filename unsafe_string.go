package contestio

import "unsafe"

func _unsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
