package contestio

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func Test_nextToken(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		bufSize   int
		wantToken string
		wantErr   error
	}{
		{
			name:      "simple token",
			input:     "abc 123",
			bufSize:   32,
			wantToken: "abc",
			wantErr:   nil,
		},
		{
			name:      "token at EOF",
			input:     "hello",
			bufSize:   32,
			wantToken: "hello",
			wantErr:   io.EOF,
		},
		{
			name:      "token exactly max size",
			input:     strings.Repeat("x", 31) + " y",
			bufSize:   32,
			wantToken: strings.Repeat("x", 31),
			wantErr:   nil,
		},
		{
			name:      "token exactly buf size",
			input:     strings.Repeat("x", 32) + " y",
			bufSize:   32,
			wantToken: strings.Repeat("x", 32),
			wantErr:   ErrTokenTooLong,
		},
		{
			name:      "token exceeds buf size",
			input:     strings.Repeat("x", 33) + " y",
			bufSize:   32,
			wantToken: strings.Repeat("x", 32),
			wantErr:   ErrTokenTooLong,
		},
		{
			name:      "leading spaces (token empty)",
			input:     "   abc",
			bufSize:   32,
			wantToken: "", // nextToken не пропускает пробелы, поэтому вернёт пустой токен
			wantErr:   nil,
		},
		{
			name:      "only spaces (EOF after spaces)",
			input:     "   ",
			bufSize:   32,
			wantToken: "",
			wantErr:   nil, // ошибки нет, возвращается пустой токен.
		},
		{
			name:      "empty input",
			input:     "",
			bufSize:   32,
			wantToken: "",
			wantErr:   io.EOF,
		},
		{
			name:      "token followed by newline",
			input:     "token\n",
			bufSize:   32,
			wantToken: "token",
			wantErr:   nil,
		},
		{
			name:      "token with spaces inside",
			input:     "abc def", // второй токен нас не интересует
			bufSize:   32,
			wantToken: "abc",
			wantErr:   nil,
		},
		{
			name:      "unicode? but bytewise, fine",
			input:     "привет",
			bufSize:   32,
			wantToken: "привет", // каждый символ занимает 2 байта
			wantErr:   io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			br := NewReaderSize(r, tt.bufSize)
			token, err := nextToken(br)

			// проверяем токен и ошибку
			if string(token) != tt.wantToken {
				t.Errorf("nextToken token = %q, want %q", token, tt.wantToken)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("nextToken error = %v, want %v", err, tt.wantErr)
			}

			remain, _ := io.ReadAll(br) // читаем всё, что осталось (весь буфер)
			wantRemain := tt.input[len(tt.wantToken):]
			if string(remain) != wantRemain {
				t.Errorf("after call, remaining = %q, want %q", remain, wantRemain)
			}
		})
	}
}

func Test_skipSpace(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		bufSize       int
		stopAtNewLine bool
		wantErr       error
		wantPos       int // expected position after SkipSpace (byte index)
	}{
		{"spaces", "  abc", 16, false, nil, 2},
		{"no spaces", "abc", 16, false, nil, 0},
		{"empty", "", 16, false, io.EOF, 0},
		{"only spaces", "   ", 16, false, io.EOF, 3},
		{"different spaces", "\t\r\n abc", 16, false, nil, 4},
		{"over buffer", "                    abc", 16, false, nil, 20},

		// stop at new line
		{"spaces then newline", "  \nx", 16, true, EOL, 3},
		{"newline only", "\nx", 16, true, EOL, 1},
		{"over buffer newline", "                    \nx", 16, true, EOL, 21},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReaderSize(strings.NewReader(tt.input), tt.bufSize)
			err := skipSpace(br, tt.stopAtNewLine)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SkipSpace() error = %v, wantErr %v", err, tt.wantErr)
			}
			// check position by reading next byte (if any)
			if tt.wantErr == nil {
				c, _ := br.ReadByte()
				if c != tt.input[tt.wantPos] {
					t.Errorf("after SkipSpace, next byte = %q, want %q", c, tt.input[tt.wantPos])
				}
			}
		})
	}
}

func Test_skipToNewLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		bufSize int
		wantErr error
		wantPos int // expected position after SkipSpace (byte index)
	}{
		{"simple", "# abc\nx", 16, nil, 6},
		{"spaces after lf", "# abc\n x", 16, nil, 6},
		{"only lf", "\nx", 16, nil, 1},
		{"crlf", "\r\nx", 16, nil, 2},
		{"lfcr", "\n\r", 16, nil, 1},
		{"empty", "", 16, io.EOF, 0},
		{"over buffer", "                    abc\nx", 16, nil, 24},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReaderSize(strings.NewReader(tt.input), tt.bufSize)
			err := skipToNewLine(br)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SkipSpace() error = %v, wantErr %v", err, tt.wantErr)
			}
			// check position by reading next byte (if any)
			if tt.wantErr == nil {
				c, _ := br.ReadByte()
				if c != tt.input[tt.wantPos] {
					t.Errorf("after SkipSpace, next byte = %q, want %q", c, tt.input[tt.wantPos])
				}
			}
		})
	}
}
