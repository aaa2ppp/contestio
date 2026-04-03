package contestio

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func Test_scanBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		delim   byte
		want    []byte
		wantErr error
	}{
		{
			name:    "simple line with newline",
			input:   "hello world\n",
			delim:   '\n',
			want:    []byte("hello world"),
			wantErr: nil,
		},
		{
			name:    "line with trailing spaces and newline",
			input:   "hello world  \t \n",
			delim:   '\n',
			want:    []byte("hello world"),
			wantErr: nil,
		},
		{
			name:    "line with spaces before newline",
			input:   "hello world  \n  ",
			delim:   '\n',
			want:    []byte("hello world"),
			wantErr: nil,
		},
		{
			name:    "no newline, eof after data",
			input:   "hello world",
			delim:   '\n',
			want:    []byte("hello world"),
			wantErr: nil,
		},
		{
			name:    "no newline, eof after data with trailing spaces",
			input:   "hello world  \t ",
			delim:   '\n',
			want:    []byte("hello world"),
			wantErr: nil,
		},
		{
			name:    "empty input",
			input:   "",
			delim:   '\n',
			want:    nil,
			wantErr: io.EOF,
		},
		{
			name:    "only spaces",
			input:   "   \t  ",
			delim:   '\n',
			want:    []byte{}, // после обрезки пробелов остаётся пустой срез
			wantErr: nil,
		},
		{
			name:    "newline only",
			input:   "\n",
			delim:   '\n',
			want:    []byte{},
			wantErr: nil,
		},
		{
			name:    "newline with spaces after",
			input:   "\n   ",
			delim:   '\n',
			want:    []byte{},
			wantErr: nil,
		},
		{
			name:    "multiple lines, read first",
			input:   "first line\nsecond line\n",
			delim:   '\n',
			want:    []byte("first line"),
			wantErr: nil,
		},
		{
			name:    "delimiter not found, read all",
			input:   "first line\nsecond line",
			delim:   'x',
			want:    []byte("first line\nsecond line"),
			wantErr: nil,
		},
		{
			name:    "comma delimiter with spaces",
			input:   "42  ,x",
			delim:   ',',
			want:    []byte("42"),
			wantErr: nil,
		},
		{
			name:    "comma delimiter without spaces",
			input:   "42,x",
			delim:   ',',
			want:    []byte("42"),
			wantErr: nil,
		},
		{
			name:    "comma delimiter at end of file",
			input:   "42,",
			delim:   ',',
			want:    []byte("42"),
			wantErr: nil,
		},
		{
			name:    "comma delimiter with trailing newline",
			input:   "42,\n",
			delim:   ',',
			want:    []byte("42"),
			wantErr: nil,
		},
		{
			name:    "only comma",
			input:   ",",
			delim:   ',',
			want:    []byte{},
			wantErr: nil,
		},
		{
			name:    "comma and spaces only",
			input:   ",  ",
			delim:   ',',
			want:    []byte{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			got, err := _scanBytes(r, tt.delim)

			if err != tt.wantErr {
				t.Errorf("ReadBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("ReadBytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadBytesMultipleCalls(t *testing.T) {
	input := "first line\nsecond line\nthird line"
	r := NewReader(strings.NewReader(input))

	// Первая строка с \n
	got, err := _scanBytes(r, '\n')
	if err != nil {
		t.Errorf("first call error: %v", err)
	}
	if string(got) != "first line" {
		t.Errorf("first call = %q, want %q", got, "first line")
	}

	// Вторая строка с \n
	got, err = _scanBytes(r, '\n')
	if err != nil {
		t.Errorf("second call error: %v", err)
	}
	if string(got) != "second line" {
		t.Errorf("second call = %q, want %q", got, "second line")
	}

	// Третья строка без завершающего \n (EOF после данных)
	got, err = _scanBytes(r, '\n')
	if err != nil {
		t.Errorf("third call error: %v, want nil", err)
	}
	if string(got) != "third line" {
		t.Errorf("third call = %q, want %q", got, "third line")
	}

	// Четвёртый вызов: уже ничего нет
	got, err = _scanBytes(r, '\n')
	if err != io.EOF {
		t.Errorf("fourth call error = %v, want io.EOF", err)
	}
	if got != nil {
		t.Errorf("fourth call = %q, want nil", got)
	}
}

func TestReadBytesOnlySpacesThenEOF(t *testing.T) {
	input := "   \t\n"
	r := NewReader(strings.NewReader(input))

	// Читаем строку с разделителем: после обрезки должно быть пусто
	got, err := _scanBytes(r, '\n')
	if err != nil {
		t.Errorf("first call error: %v", err)
	}
	if string(got) != "" {
		t.Errorf("first call = %q, want %q", got, "")
	}

	// Больше ничего нет
	got, err = _scanBytes(r, '\n')
	if err != io.EOF {
		t.Errorf("second call error = %v, want io.EOF", err)
	}
	if got != nil {
		t.Errorf("second call = %q, want nil", got)
	}
}

func TestReadBytesTrailingSpacesAfterDelimiter(t *testing.T) {
	input := "data  \n  next"
	r := NewReader(strings.NewReader(input))

	// Читаем первую строку с разделителем
	got, err := _scanBytes(r, '\n')
	if err != nil {
		t.Errorf("first call error: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("first call = %q, want %q", got, "data")
	}

	// Читаем остаток до конца (разделитель не встретится)
	got, err = _scanBytes(r, '\n')
	if err != nil {
		t.Errorf("second call error: %v, want nil", err)
	}
	if string(got) != "  next" { // don't trim the left spaces
		t.Errorf("second call = %q, want %q", got, "  next")
	}

	got, err = _scanBytes(r, '\n')
	if err != io.EOF {
		t.Errorf("third call error = %v, want io.EOF", err)
	}
	if got != nil {
		t.Errorf("third call = %q, want nil", got)
	}
}
