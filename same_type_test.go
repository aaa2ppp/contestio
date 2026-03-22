package contestio

import (
	"reflect"
	"strings"
	"testing"
)

// Тесты на сохранение типа для публичных функций ScanXXXsLn
type MyString string
type MyStringSlice []MyString

type MyInt int
type MyIntSlice []MyInt

type MyFloat float64
type MyFloatSlice []MyFloat

func TestScanSliceLn_returnsSameType(t *testing.T) {
	input := "1 2 3\n"
	br := NewReader(strings.NewReader(input))

	var s Ints[int32]
	result, err := ScanSliceLn(br, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(s) {
		t.Errorf("result type = %T, want %T", result, s)
	}
	expected := Ints[int32]{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v, want %v", result, expected)
	}
}

func TestScanWordsLn_returnsSameType(t *testing.T) {
	input := "foo bar baz\n"
	br := NewReader(strings.NewReader(input))

	var s MyStringSlice
	result, err := ScanWordsLn(br, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(s) {
		t.Errorf("result type = %T, want %T", result, s)
	}
	expected := MyStringSlice{"foo", "bar", "baz"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v, want %v", result, expected)
	}
}

func TestScanIntsLn_returnsSameType(t *testing.T) {
	input := "1 2 3\n"
	br := NewReader(strings.NewReader(input))

	var s MyIntSlice
	result, err := ScanIntsLn(br, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(s) {
		t.Errorf("result type = %T, want %T", result, s)
	}
	expected := MyIntSlice{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v, want %v", result, expected)
	}
}

func TestScanFloatsLn_returnsSameType(t *testing.T) {
	input := "1.1 2.2 3.3\n"
	br := NewReader(strings.NewReader(input))

	var s MyFloatSlice
	result, err := ScanFloatsLn(br, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(s) {
		t.Errorf("result type = %T, want %T", result, s)
	}
	expected := MyFloatSlice{1.1, 2.2, 3.3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v, want %v", result, expected)
	}
}
