//go:build any && unsafe

package contestio

import "testing"

func Test_getAnyPointerValue(t *testing.T) {
	checkGetAnyPointer(t, int(-42))
	checkGetAnyPointer(t, int8(-123))
	checkGetAnyPointer(t, int16(-12345))
	checkGetAnyPointer(t, int32(-1234567890))
	checkGetAnyPointer(t, int64(-1234567890123456789))
	checkGetAnyPointer(t, uint(42))
	checkGetAnyPointer(t, uint8(123))
	checkGetAnyPointer(t, uint16(12345))
	checkGetAnyPointer(t, uint32(1234567890))
	checkGetAnyPointer(t, uint64(1234567890123456789))
	checkGetAnyPointer(t, float32(3.1415927))
	checkGetAnyPointer(t, float64(3.141592653589793))
	checkGetAnyPointer(t, "Hello, 世界")
	checkGetAnyPointer(t, MyInt(69))
	checkGetAnyPointer(t, MyFloat(2.718281828459045))
	checkGetAnyPointer(t, MyString("Don't Worry, Be Happy"))

}

func checkGetAnyPointer[T comparable](t *testing.T, val T) {
	t.Helper()

	anyPtr := any(&val)
	anyVal := any(val)

	if gotPtr := getAnyPointer[T](anyPtr); gotPtr != &val {
		t.Errorf("getAnyPointer: fail get pointer for type %T: gotPtr = %p, want %p", val, gotPtr, &val)
	}
	if gotVal := *getAnyPointer[T](anyVal); gotVal != val {
		t.Errorf("getAnyPointer: fail get value for type %T: gotVal = %val, want %val", val, gotVal, val)
	}
}
