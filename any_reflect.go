//go:build any && !unsafe

package contestio

import "reflect"

func getAnyPointer[T any](x any) *T { return (*T)(reflect.ValueOf(x).UnsafePointer()) }

func printAnyInt(bw *Writer, x any) error { // x must be any signed integer value or pointer
	var v int64
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		v = reflect.ValueOf(x).Elem().Int()
	} else {
		v = reflect.ValueOf(x).Int()
	}
	return printInt(bw, v)
}

func printAnyUint(bw *Writer, x any) error { // x must be any unsigned integer value or pointer
	var v uint64
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		v = reflect.ValueOf(x).Elem().Uint()
	} else {
		v = reflect.ValueOf(x).Uint()
	}
	return printInt(bw, v)
}

func printAnyFloat[T Float](bw *Writer, x any) error { // x must be Float value or pointer
	var v float64
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		v = reflect.ValueOf(x).Elem().Float()
	} else {
		v = reflect.ValueOf(x).Float()
	}
	return printFloat(bw, T(v))
}

func printAnyString(bw *Writer, x any) error { // x must be any string value or pointer
	var v string
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		v = reflect.ValueOf(x).Elem().String()
	} else {
		v = reflect.ValueOf(x).String()
	}
	return printWord(bw, v)
}

var printAnyTab = []printAnyFunc{
	reflect.Int:     printAnyInt,
	reflect.Int8:    printAnyInt,
	reflect.Int16:   printAnyInt,
	reflect.Int32:   printAnyInt,
	reflect.Int64:   printAnyInt,
	reflect.Uint:    printAnyUint,
	reflect.Uint8:   printAnyUint,
	reflect.Uint16:  printAnyUint,
	reflect.Uint32:  printAnyUint,
	reflect.Uint64:  printAnyUint,
	reflect.Uintptr: printAnyUint,
	reflect.Float32: printAnyFloat[float32],
	reflect.Float64: printAnyFloat[float64],
	reflect.String:  printAnyString,
}
