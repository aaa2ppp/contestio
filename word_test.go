package contestio

import (
	"bytes"
	"testing"
)

func Test_printWordSlice(t *testing.T) {
	type testCase[T ~string] struct {
		name    string
		opts    _writeOpts
		a       []T
		wantOut string
		wantN   int
		wantErr error
	}
	tests := []testCase[string]{
		{
			name:    "empty slice",
			opts:    _writeOpts{Begin: "[", Sep: ", ", End: "]"},
			a:       []string{},
			wantOut: "[]",
			wantN:   0,
			wantErr: nil,
		},
		{
			name:    "single element",
			opts:    _writeOpts{Begin: "[", Sep: ", ", End: "]"},
			a:       []string{"foo"},
			wantOut: "[foo]",
			wantN:   1,
			wantErr: nil,
		},
		{
			name:    "multiple elements",
			opts:    _writeOpts{Begin: "[", Sep: ", ", End: "]"},
			a:       []string{"foo", "bar", "baz"},
			wantOut: "[foo, bar, baz]",
			wantN:   3,
			wantErr: nil,
		},
		{
			name:    "custom separators",
			opts:    _writeOpts{Begin: "(", Sep: "|", End: ")"},
			a:       []string{"foo", "bar", "baz"},
			wantOut: "(foo|bar|baz)",
			wantN:   3,
			wantErr: nil,
		},
		{
			name:    "empty begin/end",
			opts:    _writeOpts{Begin: "", Sep: ", ", End: ""},
			a:       []string{"foo", "bar", "baz"},
			wantOut: "foo, bar, baz",
			wantN:   3,
			wantErr: nil,
		},
		{
			name:    "empty sep",
			opts:    _writeOpts{Begin: "[", Sep: "", End: "]"},
			a:       []string{"foo", "bar", "baz"},
			wantOut: "[foobarbaz]",
			wantN:   3,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			bw := NewWriter(buf)
			gotN, gotErr := _printSliceCommon(bw, tt.opts, _printString, tt.a)
			if gotErr != tt.wantErr {
				t.Errorf("error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != tt.wantN {
				t.Errorf("n = %d, want %d", gotN, tt.wantN)
			}
			bw.Flush()
			if gotOut := buf.String(); gotOut != tt.wantOut {
				t.Errorf("output = %q, want %q", gotOut, tt.wantOut)
			}
		})
	}
}
