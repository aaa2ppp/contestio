//go:build any && unsafe

package contestio

import (
	"reflect"
	"unsafe"
)

// eface - the raw interface{} structure.
//
// See:
// file://$GOROOT/src/runtime/runtime2.go
// https://github.com/golang/go/blob/master/src/runtime/runtime2.go
type eface struct {
	_type *reflect.Type
	data  unsafe.Pointer
}

func getAnyPointer[T any](x any) *T {
	_ = [1]byte{}[unsafe.Sizeof(eface{})-unsafe.Sizeof(any(nil))] // check eface struct size in compile time
	return (*T)((*(*eface)(unsafe.Pointer(&x))).data)
}

func getAnyValue[T any](x any) T { return *getAnyPointer[T](x) }

func parseAnyInt[T Int](token []byte, x any) error { // x must be Int pointer
	v, err := parseInt[T](token)
	if err != nil {
		return err
	}
	p := getAnyPointer[T](x)
	*p = v
	return nil
}

func parseAnyFloat[T Float](token []byte, x any) error { // x must be Float pointer
	v, err := parseFloat[T](token)
	if err != nil {
		return err
	}
	p := getAnyPointer[T](x)
	*p = v
	return nil
}

func parseAnyWord[T ~string](token []byte, x any) error { // x must be ~string pointer
	v, err := parseWord[T](token)
	if err != nil {
		return err
	}
	p := getAnyPointer[T](x)
	*p = v
	return nil
}

var parseAnyTab = []parseAnyFunc{
	reflect.Int:     parseAnyInt[int],
	reflect.Int8:    parseAnyInt[int8],
	reflect.Int16:   parseAnyInt[int16],
	reflect.Int32:   parseAnyInt[int32],
	reflect.Int64:   parseAnyInt[int64],
	reflect.Uint:    parseAnyInt[uint],
	reflect.Uint8:   parseAnyInt[uint8],
	reflect.Uint16:  parseAnyInt[uint16],
	reflect.Uint32:  parseAnyInt[uint32],
	reflect.Uint64:  parseAnyInt[uint64],
	reflect.Uintptr: parseAnyInt[uintptr],
	reflect.Float32: parseAnyFloat[float32],
	reflect.Float64: parseAnyFloat[float64],
	reflect.String:  parseAnyWord[string],
}

func appendAnyInt[T Int](b []byte, x any) []byte { // x must be Int value
	v := getAnyValue[T](x)
	return appendInt(b, v)
}

func appendAnyFloat[T Float](b []byte, x any) []byte { // x must be Float value
	v := getAnyValue[T](x)
	return appendFloat(b, v)
}

var appendAnyTab = []appendAnyFunc{
	reflect.Int:     appendAnyInt[int],
	reflect.Int8:    appendAnyInt[int8],
	reflect.Int16:   appendAnyInt[int16],
	reflect.Int32:   appendAnyInt[int32],
	reflect.Int64:   appendAnyInt[int64],
	reflect.Uint:    appendAnyInt[uint],
	reflect.Uint8:   appendAnyInt[uint8],
	reflect.Uint16:  appendAnyInt[uint16],
	reflect.Uint32:  appendAnyInt[uint32],
	reflect.Uint64:  appendAnyInt[uint64],
	reflect.Uintptr: appendAnyInt[uintptr],
	reflect.Float32: appendAnyFloat[float32],
	reflect.Float64: appendAnyFloat[float64],
}

func getAnyString(x any) string { return getAnyValue[string](x) }
