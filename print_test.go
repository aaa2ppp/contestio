package contestio

import (
	"strings"
	"testing"
)

func Test_printIntSlice(t *testing.T) {
	sliceWO := _writeOpts{Begin: "[]int{", Sep: ", ", End: "}"}

	tests := []struct {
		name    string
		bufSize int
		op      _writeOpts
		a       []int
		wantOut string
		wantErr bool
	}{
		{
			"simple",
			16,
			sliceWO,
			[]int{1, 2, 3, 4, 5, 6},
			"[]int{1, 2, 3, 4, 5, 6}",
			false,
		},
		{
			"one",
			16,
			sliceWO,
			[]int{1},
			"[]int{1}",
			false,
		},
		{
			"empty",
			16,
			sliceWO,
			[]int{},
			"[]int{}",
			false,
		},
		{
			"nil",
			16,
			sliceWO,
			nil,
			"[]int{}",
			false,
		},
		{
			"over buffer",
			16,
			sliceWO,
			make([]int, 100),
			"[]int{" + strings.Repeat("0, ", 100)[:298] + "}",
			false,
		},
		{
			"zero opts",
			16,
			_writeOpts{},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			"1234567890",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			bw := NewWriterSize(&out, tt.bufSize)

			gotN, gotErr := _printSliceCommon(bw, tt.op, _printInt, tt.a)
			bw.Flush()

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("printSlice() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("printSlice() succeeded unexpectedly")
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("printSlice() out = %v, want %v", gotOut, tt.wantOut)
			}
			if gotN != len(tt.a) {
				t.Errorf("printSlice() = %v, want %v", gotN, len(tt.a))
			}
		})
	}
}
