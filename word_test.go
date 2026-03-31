package contestio

import (
	"bytes"
	"testing"
)

func Test_printWordSlice(t *testing.T) {
	type testCase[T ~string] struct {
		name    string
		opts    writeOpts
		a       []T
		wantOut string
		wantN   int
		wantErr error
	}
	tests := []testCase[string]{
		{
			name:    "empty slice",
			opts:    writeOpts{Begin: "[", Sep: ", ", End: "]"},
			a:       []string{},
			wantOut: "[]",
			wantN:   0,
			wantErr: nil,
		},
		{
			name:    "single element",
			opts:    writeOpts{Begin: "[", Sep: ", ", End: "]"},
			a:       []string{"foo"},
			wantOut: "[foo]",
			wantN:   1,
			wantErr: nil,
		},
		{
			name:    "multiple elements",
			opts:    writeOpts{Begin: "[", Sep: ", ", End: "]"},
			a:       []string{"foo", "bar", "baz"},
			wantOut: "[foo, bar, baz]",
			wantN:   3,
			wantErr: nil,
		},
		{
			name:    "custom separators",
			opts:    writeOpts{Begin: "(", Sep: "|", End: ")"},
			a:       []string{"foo", "bar", "baz"},
			wantOut: "(foo|bar|baz)",
			wantN:   3,
			wantErr: nil,
		},
		{
			name:    "empty begin/end",
			opts:    writeOpts{Begin: "", Sep: ", ", End: ""},
			a:       []string{"foo", "bar", "baz"},
			wantOut: "foo, bar, baz",
			wantN:   3,
			wantErr: nil,
		},
		{
			name:    "empty sep",
			opts:    writeOpts{Begin: "[", Sep: "", End: "]"},
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
			gotN, gotErr := printSliceCommon(bw, tt.opts, printString, tt.a)
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
