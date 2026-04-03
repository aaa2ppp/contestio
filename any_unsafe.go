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
	_type unsafe.Pointer
	data  unsafe.Pointer
}

func _getAnyPointer[T any](x any) *T {
	_ = [1]byte{}[unsafe.Sizeof(eface{})-unsafe.Sizeof(any(nil))] // check eface struct size in compile time
	return (*T)((*(*eface)(unsafe.Pointer(&x))).data)
}

func _printAnyInt[T Int](bw *Writer, x any) error { // x must be Int value or pointer
	p := _getAnyPointer[T](x)
	return _printInt(bw, *p)
}

func _printAnyFloat[T Float](bw *Writer, x any) error { // x must be Float value or pointer
	p := _getAnyPointer[T](x)
	return _printFloat(bw, *p)
}

func _printAnyString[T ~string](bw *Writer, x any) error { // x must be any string value or pointer
	p := _getAnyPointer[T](x)
	return _printString(bw, *p)
}

var _printAnyTab = []_printAnyFunc{
	reflect.Int:     _printAnyInt[int],
	reflect.Int8:    _printAnyInt[int8],
	reflect.Int16:   _printAnyInt[int16],
	reflect.Int32:   _printAnyInt[int32],
	reflect.Int64:   _printAnyInt[int64],
	reflect.Uint:    _printAnyInt[uint],
	reflect.Uint8:   _printAnyInt[uint8],
	reflect.Uint16:  _printAnyInt[uint16],
	reflect.Uint32:  _printAnyInt[uint32],
	reflect.Uint64:  _printAnyInt[uint64],
	reflect.Uintptr: _printAnyInt[uintptr],
	reflect.Float32: _printAnyFloat[float32],
	reflect.Float64: _printAnyFloat[float64],
	reflect.String:  _printAnyString[string],
}
