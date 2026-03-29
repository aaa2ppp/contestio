//go:build any && !unsafe

package contestio

import "reflect"

func getAnyPointer[T any](x any) *T { return (*T)(reflect.ValueOf(x).UnsafePointer()) }

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

func appendAnyInt(b []byte, x any) []byte { // x must be any signed int value or pointer
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		return appendInt(b, reflect.ValueOf(x).Elem().Int())
	}
	return appendInt(b, reflect.ValueOf(x).Int())
}

func appendAnyUint(b []byte, x any) []byte { // x must be any unsigned int value or pointer
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		return appendInt(b, reflect.ValueOf(x).Elem().Uint())
	}
	return appendInt(b, reflect.ValueOf(x).Uint())
}

func appendAnyFloat[T Float](b []byte, x any) []byte { // x must be any float value or pointer
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		return appendFloat[T](b, T(reflect.ValueOf(x).Elem().Float()))
	}
	return appendFloat[T](b, T(reflect.ValueOf(x).Float()))
}

var appendAnyTab = []appendAnyFunc{
	reflect.Int:     appendAnyInt,
	reflect.Int8:    appendAnyInt,
	reflect.Int16:   appendAnyInt,
	reflect.Int32:   appendAnyInt,
	reflect.Int64:   appendAnyInt,
	reflect.Uint:    appendAnyUint,
	reflect.Uint8:   appendAnyUint,
	reflect.Uint16:  appendAnyUint,
	reflect.Uint32:  appendAnyUint,
	reflect.Uint64:  appendAnyUint,
	reflect.Uintptr: appendAnyUint,
	reflect.Float32: appendAnyFloat[float32],
	reflect.Float64: appendAnyFloat[float64],
}

func getAnyString(x any) string { // x must be string value or pointer
	if reflect.TypeOf(x).Kind() == reflect.Pointer {
		return reflect.ValueOf(x).Elem().String()
	}
	return reflect.ValueOf(x).String()
}
