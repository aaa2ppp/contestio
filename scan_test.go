package contestio

import (
	"errors"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func godParser(b []byte) (string, error)    { return string(b), nil }
func godParserTo(b []byte, p *string) error { return parseValTo(b, godParser, p) }

var parseError = errors.New("parse error")

func badParser(failAt int) parseFunc[string] {
	count := 0
	return func(b []byte) (string, error) {
		count++
		if count == failAt {
			return "", parseError
		}
		return godParser(b)
	}
}

func badParserTo(failAt int) parseToFunc[*string] {
	count := 0
	return func(b []byte, p *string) error {
		count++
		if count == failAt {
			return parseError
		}
		return godParserTo(b, p)
	}
}

func Test_scanSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		parse   func([]byte) (string, error)
		a       []string
		wantA   []string
		wantN   int
		wantErr error
	}{
		{
			"empty input",
			"",
			godParser,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			io.EOF,
		},
		{
			"only spaces",
			" \t\r\n",
			godParser,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			io.EOF,
		},
		{
			"one token",
			"one",
			godParser,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"leading spaces",
			" \t\r\none",
			godParser,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"trailing spaces",
			"one \t\r\n",
			godParser,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"correct input",
			"one two three four five six",
			godParser,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "four", "five"},
			5,
			nil,
		},
		{
			"incomplete input",
			"one two three",
			godParser,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "4", "5"},
			3,
			io.ErrUnexpectedEOF,
		},
		{
			"parse error at 1",
			"one two three four five six",
			badParser(1),
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			parseError,
		},
		{
			"parse error at 4",
			"one two three four five six",
			badParser(4),
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "4", "5"},
			3,
			parseError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			gotA := slices.Clone(tt.a)
			gotN, gotErr := scanSliceCommon(br, tt.parse, gotA)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("scanSliceCommon() error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != tt.wantN {
				t.Errorf("scanSliceCommon() n = %v, want %v", gotN, tt.wantN)
			}
			if !reflect.DeepEqual(gotA, tt.wantA) {
				t.Errorf("scanSliceCommon() out = %q, want %q", gotA, tt.wantA)
			}
		})
	}
}

func Test_scanSliceLn(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		parse   func([]byte) (string, error)
		a       []string
		n       int
		wantA   []string
		wantErr error
	}{
		{
			"empty input",
			"",
			godParser,
			nil,
			0,
			nil,
			io.EOF,
		},
		{
			"only spaces (except LF)",
			" \t\r",
			godParser,
			nil,
			0,
			nil,
			io.EOF,
		},
		{
			"only LF",
			"\n",
			godParser,
			nil,
			0,
			nil,
			nil,
		},
		{
			"only spaces ended with LF",
			" \t\r\n",
			godParser,
			nil,
			0,
			nil,
			nil,
		},
		{
			"one token (EOL)",
			"one\n",
			godParser,
			nil,
			0,
			[]string{"one"},
			nil,
		},
		{
			"one token (EOF)",
			"one",
			godParser,
			nil,
			0,
			[]string{"one"},
			nil,
		},
		{
			"leading spaces",
			" \t\rone\n",
			godParser,
			nil,
			0,
			[]string{"one"},
			nil,
		},
		{
			"trailing spaces",
			"one \t\r\n",
			godParser,
			nil,
			0,
			[]string{"one"},
			nil,
		},
		{
			"trailing spaces (EOF)",
			"one \t\r",
			godParser,
			nil,
			0,
			[]string{"one"},
			nil,
		},
		{
			"correct input",
			"one two three four five\nsix",
			godParser,
			nil,
			0,
			[]string{"one", "two", "three", "four", "five"},
			nil,
		},
		{
			"correct input (append)",
			"three four five\nsix",
			godParser,
			[]string{"1", "2", "3", "4", "5", "6"},
			2,
			[]string{"1", "2", "three", "four", "five"},
			nil,
		},
		{
			"parse error at 1",
			"one two three four five\nsix",
			badParser(1),
			nil,
			0,
			nil,
			parseError,
		},
		{
			"parse error at 4",
			"one two three four five\nsix",
			badParser(4),
			nil,
			0,
			[]string{"one", "two", "three"},
			parseError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			var a []string
			if tt.a != nil {
				a = slices.Clone(tt.a)
			}
			gotA, gotErr := scanSliceLnCommon(br, tt.parse, a[:tt.n])
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("scanSliceLnCommon() error = %v, want %v", gotErr, tt.wantErr)
			}
			if !reflect.DeepEqual(gotA, tt.wantA) {
				t.Errorf("scanSliceLnCommon() = %q, want %q", gotA, tt.wantA)
			}
			if n := len(gotA); n > 0 && n <= len(a) {
				if &a[0] != &gotA[0] {
					t.Errorf("scanSliceLnCommon(): expected same slice")
				}
			}
		})
	}
}

func Test_scanVars(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		stopAtEol bool
		parse     parseToFunc[*string]
		a         []string
		wantA     []string
		wantN     int
		wantErr   error
	}{
		{
			"empty input",
			"",
			false,
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			io.EOF,
		},
		{
			"only spaces",
			" \t\r\n",
			false,
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			io.EOF,
		},
		{
			"one token",
			"one",
			false,
			godParserTo,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"leading spaces",
			" \t\r\none",
			false,
			godParserTo,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"trailing spaces",
			"one \t\r\n",
			false,
			godParserTo,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"correct input",
			"one two three four five six",
			false,
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "four", "five"},
			5,
			nil,
		},
		{
			"incomplete input",
			"one two three",
			false,
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "4", "5"},
			3,
			io.ErrUnexpectedEOF,
		},
		{
			"incomplete input (stop at eol)",
			"one two three\nfour five six",
			true, // stop at eol
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "4", "5"},
			3,
			EOL,
		},
		{
			"parse error at 1",
			"one two three four five six",
			false,
			badParserTo(1),
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			parseError,
		},
		{
			"parse error at 4",
			"one two three four five six",
			false,
			badParserTo(4),
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "4", "5"},
			3,
			parseError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			gotA := slices.Clone(tt.a)
			p := make([]*string, len(gotA))
			for i := range gotA {
				p[i] = &gotA[i]
			}
			gotN, gotErr := scanVarsCommon(br, tt.stopAtEol, tt.parse, p)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("scanVarsCommon() error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != tt.wantN {
				t.Errorf("scanVarsCommon() n = %v, want %v", gotN, tt.wantN)
			}
			if !reflect.DeepEqual(gotA, tt.wantA) {
				t.Errorf("scanVarsCommon() out = %q, want %q", gotA, tt.wantA)
			}
		})
	}
}

func Test_scanVarsLn(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		parse   parseToFunc[*string]
		a       []string
		wantA   []string
		wantN   int
		wantErr error
	}{
		{
			"empty input",
			"",
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			io.EOF,
		},
		{
			"only spaces (except LF)",
			" \t\r",
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			io.EOF,
		},
		{
			"only LF",
			"\n",
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			EOL,
		},
		{
			"only spaces ended with LF",
			" \t\r\n",
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			EOL,
		},
		{
			"one token (EOL)",
			"one\n",
			godParserTo,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"one token (EOF)",
			"one",
			godParserTo,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"leading spaces",
			" \t\rone",
			godParserTo,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"trailing spaces (EOL)",
			"one \t\r\n",
			godParserTo,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"trailing spaces (EOF)",
			"one \t\r",
			godParserTo,
			[]string{"1"},
			[]string{"one"},
			1,
			nil,
		},
		{
			"correct input",
			"one two three four five\nsix",
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "four", "five"},
			5,
			nil,
		},
		{
			"incomplete input (EOF)",
			"one two three",
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "4", "5"},
			3,
			io.ErrUnexpectedEOF,
		},
		{
			"incomplete input (EOL)",
			"one two three\nfour five six",
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "4", "5"},
			3,
			EOL,
		},
		{
			"expected eol",
			"one two three four five six",
			godParserTo,
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "four", "five"},
			5,
			ErrExpectedEOL,
		},
		{
			"parse error at 1",
			"one two three four five\nsix",
			badParserTo(1),
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			parseError,
		},
		{
			"parse error at 4",
			"one two three four five\nsix",
			badParserTo(4),
			[]string{"1", "2", "3", "4", "5"},
			[]string{"one", "two", "three", "4", "5"},
			3,
			parseError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			gotA := slices.Clone(tt.a)
			p := make([]*string, len(gotA))
			for i := range gotA {
				p[i] = &gotA[i]
			}
			gotN, gotErr := scanVarsLn(br, tt.parse, p...)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("scanVarsLnCommon() error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != tt.wantN {
				t.Errorf("scanVarsLnCommon() n = %v, want %v", gotN, tt.wantN)
			}
			if !reflect.DeepEqual(gotA, tt.wantA) {
				t.Errorf("scanVarsLnCommon() out = %q, want %q", gotA, tt.wantA)
			}
		})
	}
}
