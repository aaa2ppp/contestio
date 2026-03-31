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

func getAnyPointer[T any](x any) *T {
	_ = [1]byte{}[unsafe.Sizeof(eface{})-unsafe.Sizeof(any(nil))] // check eface struct size in compile time
	return (*T)((*(*eface)(unsafe.Pointer(&x))).data)
}

func printAnyInt[T Int](bw *Writer, x any) error { // x must be Int value or pointer
	p := getAnyPointer[T](x)
	return printInt(bw, *p)
}

func printAnyFloat[T Float](bw *Writer, x any) error { // x must be Float value or pointer
	p := getAnyPointer[T](x)
	return printFloat(bw, *p)
}

func printAnyString[T ~string](bw *Writer, x any) error { // x must be any string value or pointer
	p := getAnyPointer[T](x)
	return printString(bw, *p)
}

var printAnyTab = []printAnyFunc{
	reflect.Int:     printAnyInt[int],
	reflect.Int8:    printAnyInt[int8],
	reflect.Int16:   printAnyInt[int16],
	reflect.Int32:   printAnyInt[int32],
	reflect.Int64:   printAnyInt[int64],
	reflect.Uint:    printAnyInt[uint],
	reflect.Uint8:   printAnyInt[uint8],
	reflect.Uint16:  printAnyInt[uint16],
	reflect.Uint32:  printAnyInt[uint32],
	reflect.Uint64:  printAnyInt[uint64],
	reflect.Uintptr: printAnyInt[uintptr],
	reflect.Float32: printAnyFloat[float32],
	reflect.Float64: printAnyFloat[float64],
	reflect.String:  printAnyString[string],
}
