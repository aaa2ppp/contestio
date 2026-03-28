//go:build any && unsafe

package contestio

import "testing"

func Test_getAnyPointerValue(t *testing.T) {
	checkGetAnyPointerValue(t, int(-42))
	checkGetAnyPointerValue(t, int8(-123))
	checkGetAnyPointerValue(t, int16(-12345))
	checkGetAnyPointerValue(t, int32(-1234567890))
	checkGetAnyPointerValue(t, int64(-1234567890123456789))
	checkGetAnyPointerValue(t, uint(42))
	checkGetAnyPointerValue(t, uint8(123))
	checkGetAnyPointerValue(t, uint16(12345))
	checkGetAnyPointerValue(t, uint32(1234567890))
	checkGetAnyPointerValue(t, uint64(1234567890123456789))
	checkGetAnyPointerValue(t, float32(3.1415927))
	checkGetAnyPointerValue(t, float64(3.141592653589793))
	checkGetAnyPointerValue(t, "Hello, 世界")
}

func checkGetAnyPointerValue[T comparable](t *testing.T, v T) {
	t.Helper()
	if ptr := getAnyPointer[T](any(&v)); ptr != &v {
		t.Errorf("getPointer for type %T: ptr = %p, want %p", v, ptr, &v)
	}
	if val := getAnyValue[T](any(v)); val != v {
		t.Errorf("getValue for type %T: val = %v, want %v", v, val, v)
	}
	if val := getAnyValue[T](any(&v)); val != v {
		t.Errorf("getValue for type %T: val = %v, want %v", v, val, v)
	}
}
