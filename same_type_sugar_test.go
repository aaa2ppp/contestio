//go:build sugar

package contestio

import (
	"reflect"
	"strings"
	"testing"
)

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
