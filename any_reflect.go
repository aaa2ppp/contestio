//go:build any && !unsafe

package contestio

import "reflect"

func _getAnyPointer[T any](x any) *T { return (*T)(reflect.ValueOf(x).UnsafePointer()) }

func _printAnyInt(bw *Writer, x any) error { // x must be any signed integer value or pointer
	var v int64
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		v = reflect.ValueOf(x).Elem().Int()
	} else {
		v = reflect.ValueOf(x).Int()
	}
	return _printInt(bw, v)
}

func _printAnyUint(bw *Writer, x any) error { // x must be any unsigned integer value or pointer
	var v uint64
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		v = reflect.ValueOf(x).Elem().Uint()
	} else {
		v = reflect.ValueOf(x).Uint()
	}
	return _printInt(bw, v)
}

func _printAnyFloat[T Float](bw *Writer, x any) error { // x must be Float value or pointer
	var v float64
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		v = reflect.ValueOf(x).Elem().Float()
	} else {
		v = reflect.ValueOf(x).Float()
	}
	return _printFloat(bw, T(v))
}

func _printAnyString(bw *Writer, x any) error { // x must be any string value or pointer
	var v string
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		v = reflect.ValueOf(x).Elem().String()
	} else {
		v = reflect.ValueOf(x).String()
	}
	return _printString(bw, v)
}

var _printAnyTab = []_printAnyFunc{
	reflect.Int:     _printAnyInt,
	reflect.Int8:    _printAnyInt,
	reflect.Int16:   _printAnyInt,
	reflect.Int32:   _printAnyInt,
	reflect.Int64:   _printAnyInt,
	reflect.Uint:    _printAnyUint,
	reflect.Uint8:   _printAnyUint,
	reflect.Uint16:  _printAnyUint,
	reflect.Uint32:  _printAnyUint,
	reflect.Uint64:  _printAnyUint,
	reflect.Uintptr: _printAnyUint,
	reflect.Float32: _printAnyFloat[float32],
	reflect.Float64: _printAnyFloat[float64],
	reflect.String:  _printAnyString,
}
