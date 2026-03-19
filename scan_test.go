package contestio

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func Test_scanSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		n       int
		want    []string
		wantErr error
	}{
		{
			"simple",
			"1 2 3 4 5",
			5,
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
		{
			"simple sasha",
			"Shla Sasha po shosse i sosala sushku.",
			7,
			[]string{"Shla", "Sasha", "po", "shosse", "i", "sosala", "sushku."},
			nil,
		},
		{
			"multiple spaces",
			"1  2   3    4     5",
			5,
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
		{
			"various spaces",
			"1\t\r\n2",
			2,
			[]string{"1", "2"},
			nil,
		},
		{
			"leading spaces",
			"   1",
			1,
			[]string{"1"},
			nil,
		},
		{
			"final spaces",
			"1   ",
			1,
			[]string{"1"},
			nil,
		},
		{
			"too few elements",
			"1 2 3 4 5",
			6,
			[]string{"1", "2", "3", "4", "5"},
			io.ErrUnexpectedEOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			out := make([]string, tt.n)
			gotN, gotErr := scanSlice(br, func(b []byte) (string, error) { return string(b), nil }, out)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("scanSlice() error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != len(tt.want) {
				t.Errorf("scanSlice() n = %v, want %v", gotN, len(tt.want))
			}
			if !reflect.DeepEqual(out[:min(gotN, len(out))], tt.want) {
				t.Errorf("scanSlice() out = %q, want %q", out, tt.want)
			}
		})
	}
}

func Test_scanSliceLn(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr error
	}{
		{
			"simple",
			"1 2 3 4 5\n",
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
		{
			"simple sasha",
			"Shla Sasha po shosse i sosala sushku.\n",
			[]string{"Shla", "Sasha", "po", "shosse", "i", "sosala", "sushku."},
			nil,
		},
		{
			"multiple spaces",
			"1  2   3    4     5\n",
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
		{
			"various spaces",
			"1\t\r 2\n",
			[]string{"1", "2"},
			nil,
		},
		{
			"leading spaces",
			"   1\n",
			[]string{"1"},
			nil,
		},
		{
			"final spaces",
			"1   \n",
			[]string{"1"},
			nil,
		},
		{
			"no lf",
			"1 2 3 4 5",
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			out, gotErr := scanSliceLn(br, func(b []byte) (string, error) { return string(b), nil }, nil)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("scanSliceLn() error = %v, want %v", gotErr, tt.wantErr)
			}
			if !reflect.DeepEqual(out, tt.want) {
				t.Errorf("scanSliceLn() out = %q, want %q", out, tt.want)
			}
		})
	}
}

func Test_scanVars(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		n       int
		want    []string
		wantErr error
	}{
		{
			"simple",
			"1 2 3 4 5",
			5,
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
		{
			"simple sasha",
			"Shla Sasha po shosse i sosala sushku.",
			7,
			[]string{"Shla", "Sasha", "po", "shosse", "i", "sosala", "sushku."},
			nil,
		},
		{
			"multiple spaces",
			"1  2   3    4     5",
			5,
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
		{
			"various spaces",
			"1\t\r\n2",
			2,
			[]string{"1", "2"},
			nil,
		},
		{
			"leading spaces",
			"   1",
			1,
			[]string{"1"},
			nil,
		},
		{
			"final spaces",
			"1   ",
			1,
			[]string{"1"},
			nil,
		},
		{
			"too few elements",
			"1 2 3 4 5",
			6,
			[]string{"1", "2", "3", "4", "5"},
			io.ErrUnexpectedEOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			out := make([]string, tt.n)
			vars := make([]*string, tt.n)
			for i := range out {
				vars[i] = &out[i]
			}
			gotN, gotErr := scanVars(br, func(b []byte) (string, error) { return string(b), nil }, vars...)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("scanVars() error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != len(tt.want) {
				t.Errorf("scanVars() n = %v, want %v", gotN, len(tt.want))
			}
			if !reflect.DeepEqual(out[:min(gotN, len(out))], tt.want) {
				t.Errorf("scanVars() out = %q, want %q", out, tt.want)
			}
		})
	}
}

func Test_scanVarsLn(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		n       int
		want    []string
		wantErr error
	}{
		{
			"simple",
			"1 2 3 4 5\n",
			5,
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
		{
			"simple sasha",
			"Shla Sasha po shosse i sosala sushku.\n",
			7,
			[]string{"Shla", "Sasha", "po", "shosse", "i", "sosala", "sushku."},
			nil,
		},
		{
			"multiple spaces",
			"1  2   3    4     5\n",
			5,
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
		{
			"various spaces",
			"1\t\r 2\n",
			2,
			[]string{"1", "2"},
			nil,
		},
		{
			"leading spaces",
			"   1\n",
			1,
			[]string{"1"},
			nil,
		},
		{
			"final spaces",
			"1   \n",
			1,
			[]string{"1"},
			nil,
		},
		{
			"too few elements",
			"1 2 3 4 5\n 6",
			6,
			[]string{"1", "2", "3", "4", "5"},
			EOL,
		},
		{
			"no lf",
			"1 2 3 4 5",
			5,
			[]string{"1", "2", "3", "4", "5"},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			out := make([]string, tt.n)
			vars := make([]*string, tt.n)
			for i := range out {
				vars[i] = &out[i]
			}
			gotN, gotErr := scanVarsLn(br, func(b []byte) (string, error) { return string(b), nil }, vars...)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("scanVarsLn() error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != len(tt.want) {
				t.Errorf("scanVarsLn() n = %v, want %v", gotN, len(tt.want))
			}
			if !reflect.DeepEqual(out[:min(gotN, len(out))], tt.want) {
				t.Errorf("scanVarsLn() out = %q, want %q", out, tt.want)
			}
		})
	}
}

func Test_scanSliceLn_behavior(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr error
	}{
		{"normal line with newline", "a b c\n", []string{"a", "b", "c"}, nil},
		{"no newline at end", "a b c", []string{"a", "b", "c"}, nil},
		{"empty line", "\n", nil, nil},
		{"empty input", "", nil, io.EOF},
		{"only spaces with newline", "   \n", nil, nil},
		{"only spaces no newline", "   ", nil, io.EOF},
		{"single token then EOF", "hello", []string{"hello"}, nil},
		{"multiple tokens then EOF", "hello world", []string{"hello", "world"}, nil},
		{"tokens with trailing spaces", "a b   ", []string{"a", "b"}, nil}, // trailing spaces, no newline
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			got, err := scanSliceLn(br, func(b []byte) (string, error) { return string(b), nil }, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scanVarsLn_behavior(t *testing.T) {
	parseStr := func(b []byte) (string, error) { return string(b), nil }

	tests := []struct {
		name     string
		input    string
		request  int      // сколько переменных запрашиваем
		wantVars []string // ожидаемые значения после чтения (непрочитанные останутся пустыми)
		wantN    int
		wantErr  error
	}{
		{"normal line with newline", "a b c\n", 3, []string{"a", "b", "c"}, 3, nil},
		{"no newline at end", "a b c", 3, []string{"a", "b", "c"}, 3, nil},
		{"partial data (less than requested)", "a b", 3, []string{"a", "b", ""}, 2, io.ErrUnexpectedEOF},
		{"empty line", "\n", 1, []string{""}, 0, EOL},
		{"empty input", "", 1, []string{""}, 0, io.ErrUnexpectedEOF},
		{"extra spaces after", "a b c   \n", 3, []string{"a", "b", "c"}, 3, nil},
		{"extra text after (garbage)", "a b c extra\n", 3, []string{"a", "b", "c"}, 3, ErrExpectedEOL},
		{"EOF after all tokens", "a b c", 3, []string{"a", "b", "c"}, 3, nil},
		{"more tokens than requested", "a b c d e\n", 3, []string{"a", "b", "c"}, 3, ErrExpectedEOL},
		{"request zero variables", "", 0, []string{}, 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			vars := make([]string, tt.request)
			ptrs := make([]*string, tt.request)
			for i := range vars {
				ptrs[i] = &vars[i]
			}

			gotN, err := scanVarsLn(br, parseStr, ptrs...)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if gotN != tt.wantN {
				t.Errorf("n = %d, want %d", gotN, tt.wantN)
			}
			for i := 0; i < tt.wantN; i++ {
				if vars[i] != tt.wantVars[i] {
					t.Errorf("var[%d] = %q, want %q", i, vars[i], tt.wantVars[i])
				}
			}
			// Если было запрошено больше, чем прочитано, оставшиеся должны остаться пустыми (zero value)
			for i := tt.wantN; i < tt.request; i++ {
				if vars[i] != "" {
					t.Errorf("var[%d] should be empty, got %q", i, vars[i])
				}
			}
		})
	}
}

func Test_readMultipleLines(t *testing.T) {
	input := "1 2\n3 4\n5 6"
	br := NewReader(strings.NewReader(input))
	var result [][]string

	for {
		line, err := scanSliceLn(br, func(b []byte) (string, error) { return string(b), nil }, nil)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, line)
	}

	want := [][]string{{"1", "2"}, {"3", "4"}, {"5", "6"}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func Benchmark_scanInt(b *testing.B) {
	b.StopTimer()
	N := 1 << 20 // 1M
	maxSpace := 5
	rand := rand.New(rand.NewSource(1))
	nums := generateInts[int](rand, N)
	input := makeIntsInput(rand, nums, maxSpace)
	memory := make([]int, N)

	b.Run("fmt.Fscan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			for i := range res {
				if _, err := fmt.Fscan(br, &res[i]); err != nil {
					b.Errorf("scan: %d: %v", i, err)
					return
				}
			}
		}
	})

	b.Run("ReadString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			s, err := br.ReadString('\n')
			if err != nil {
				b.Errorf("read string: %v", err)
				return
			}

			tokens := strings.Fields(s)
			if len(tokens) != N {
				b.Errorf("got tokens %d, want %d", len(tokens), N)
			}

			for i, token := range tokens {
				v, err := strconv.Atoi(token)
				if err != nil {
					b.Errorf("parse: %d: %v", i, err)
					return
				}
				res[i] = v
			}
		}
	})

	b.Run("WordScanner", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			sc := bufio.NewScanner(bytes.NewReader(input))
			sc.Buffer(make([]byte, defaultBufSize), defaultBufSize-1)
			sc.Split(bufio.ScanWords)
			res := memory[:N]
			b.StartTimer()

			for i := 0; i < len(res) && sc.Scan(); i++ {
				v, err := strconv.Atoi(sc.Text())
				if err != nil {
					b.Errorf("parse: %d: %v", i, err)
					return
				}
				res[i] = v
			}
		}
	})

	b.Run("scanSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			n, err := scanSlice(br, parseInt, res)

			if n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})

	b.Run("scanSliceLn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:0]
			b.StartTimer()

			res, err := scanSliceLn(br, parseInt, res[:0])

			if n := len(res); n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})

	b.Run("scanVars_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:0]
			b.StartTimer()

			var v1, v2, v3 int
			for i := 0; i+3 < N; i += 3 {
				n, err := scanVars(br, parseInt, &v1, &v2, &v3)
				if err != nil {
					b.Errorf("scan: %d: %d", i+n, err)
					return
				}
				res = append(res, v1, v2, v3)
			}
			n, err := scanVars(br, parseInt, &v1, &v2, &v3)
			if err != nil && err != io.EOF {
				b.Errorf("scan: %d: %d", i+n, err)
				return
			}
			switch n {
			case 1:
				res = append(res, v1)
			case 2:
				res = append(res, v1, v2)
			}

			if n := len(res); n != N || (err != nil && err != io.EOF) {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})
}

func Benchmark_scanFloat(b *testing.B) {
	b.StopTimer()
	N := 1 << 20 // 1M
	maxSpace := 5
	rand := rand.New(rand.NewSource(1))
	nums := generateFloats[float64](rand, N)
	input := makeFloatsInput(rand, nums, maxSpace)
	memory := make([]float64, N)

	b.Run("fmt.Fscan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			for i := range res {
				if _, err := fmt.Fscan(br, &res[i]); err != nil {
					b.Errorf("scan: %d: %v", i, err)
					return
				}
			}
		}
	})

	b.Run("ReadString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			s, err := br.ReadString('\n')
			if err != nil {
				b.Errorf("read string: %v", err)
				return
			}

			tokens := strings.Fields(s)
			if len(tokens) != N {
				b.Errorf("got tokens %d, want %d", len(tokens), N)
			}

			for i, token := range tokens {
				v, err := strconv.ParseFloat(token, 64)
				if err != nil {
					b.Errorf("parse: %d: %v", i, err)
					return
				}
				res[i] = v
			}
		}
	})

	b.Run("WordScanner", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			sc := bufio.NewScanner(bytes.NewReader(input))
			sc.Buffer(make([]byte, defaultBufSize), defaultBufSize-1)
			sc.Split(bufio.ScanWords)
			res := memory[:N]
			b.StartTimer()

			for i := 0; i < len(res) && sc.Scan(); i++ {
				token := sc.Text()
				v, err := strconv.ParseFloat(token, 64)
				if err != nil {
					b.Errorf("parse: %d: %v", i, err)
					return
				}
				res[i] = v
			}
		}
	})

	b.Run("scanSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			n, err := scanSlice(br, parseFloat, res)

			if n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})

	b.Run("scanSliceLn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:0]
			b.StartTimer()

			res, err := scanSliceLn(br, parseFloat, res[:0])

			if n := len(res); n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})

	b.Run("scanVars_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:0]
			b.StartTimer()

			var v1, v2, v3 float64
			for i := 0; i+3 < N; i += 3 {
				n, err := scanVars(br, parseFloat, &v1, &v2, &v3)
				if err != nil {
					b.Errorf("scan: %d: %d", i+n, err)
					return
				}
				res = append(res, v1, v2, v3)
			}
			n, err := scanVars(br, parseFloat, &v1, &v2, &v3)
			if err != nil && err != io.EOF {
				b.Errorf("scan: %d: %d", i+n, err)
				return
			}
			switch n {
			case 1:
				res = append(res, v1)
			case 2:
				res = append(res, v1, v2)
			}

			if n := len(res); n != N || (err != nil && err != io.EOF) {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})
}

func Benchmark_scanSlice(b *testing.B) {
	b.StopTimer()
	N := 1 << 20 // 1M
	maxSpace := 5
	rand := rand.New(rand.NewSource(1))
	nums := generateInts[int](rand, N)
	input := makeIntsInput(rand, nums, maxSpace)
	memory := make([]int, N)

	b.Run("parseIntStd", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			n, err := scanSlice(br, parseIntStd, res)

			if n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})
	b.Run("parseInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			n, err := scanSlice(br, parseInt, res)

			if n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})
}

func Test_scanSlice_parseError(t *testing.T) {
	// Парсер, который успешно читает "ok", но ошибается на "bad"
	parse := func(b []byte) (string, error) {
		s := string(b)
		if s == "bad" {
			return "", errors.New("parse error")
		}
		return s, nil
	}

	tests := []struct {
		name     string
		input    string
		n        int
		wantN    int
		wantVals []string
		wantErr  string
	}{
		{
			name:     "error on first token",
			input:    "bad 2 3",
			n:        3,
			wantN:    0,
			wantVals: []string{"", "", ""},
			wantErr:  "parse error",
		},
		{
			name:     "error on second token",
			input:    "1 bad 3",
			n:        3,
			wantN:    1,
			wantVals: []string{"1", "", ""},
			wantErr:  "parse error",
		},
		{
			name:     "error on last token",
			input:    "1 2 bad",
			n:        3,
			wantN:    2,
			wantVals: []string{"1", "2", ""},
			wantErr:  "parse error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			out := make([]string, tt.n)
			gotN, gotErr := scanSlice(br, parse, out)

			if gotErr == nil || gotErr.Error() != tt.wantErr {
				t.Errorf("scanSlice() error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != tt.wantN {
				t.Errorf("scanSlice() n = %d, want %d", gotN, tt.wantN)
			}
			if !reflect.DeepEqual(out, tt.wantVals) {
				t.Errorf("scanSlice() out = %v, want %v", out, tt.wantVals)
			}
		})
	}
}

func Test_scanVars_parseError(t *testing.T) {
	parse := func(b []byte) (string, error) {
		s := string(b)
		if s == "bad" {
			return "", errors.New("parse error")
		}
		return s, nil
	}

	tests := []struct {
		name     string
		input    string
		n        int
		wantN    int
		wantVals []string
		wantErr  string
	}{
		{
			name:     "error on first token",
			input:    "bad 2 3",
			n:        3,
			wantN:    0,
			wantVals: []string{"", "", ""},
			wantErr:  "parse error",
		},
		{
			name:     "error on second token",
			input:    "1 bad 3",
			n:        3,
			wantN:    1,
			wantVals: []string{"1", "", ""},
			wantErr:  "parse error",
		},
		{
			name:     "error on last token",
			input:    "1 2 bad",
			n:        3,
			wantN:    2,
			wantVals: []string{"1", "2", ""},
			wantErr:  "parse error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			out := make([]string, tt.n)
			ptrs := make([]*string, tt.n)
			for i := range out {
				ptrs[i] = &out[i]
			}
			gotN, gotErr := scanVars(br, parse, ptrs...)

			if gotErr == nil || gotErr.Error() != tt.wantErr {
				t.Errorf("scanVars() error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotN != tt.wantN {
				t.Errorf("scanVars() n = %d, want %d", gotN, tt.wantN)
			}
			if !reflect.DeepEqual(out, tt.wantVals) {
				t.Errorf("scanVars() out = %v, want %v", out, tt.wantVals)
			}
		})
	}
}

func Test_scanVarsLn_trailingSpacesWithoutNewline(t *testing.T) {
	br := NewReader(strings.NewReader("a b c   "))
	var a, b, c string
	n, err := scanVarsLn(br, func(b []byte) (string, error) { return string(b), nil }, &a, &b, &c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("n = %d, want 3", n)
	}
	if a != "a" || b != "b" || c != "c" {
		t.Errorf("got %q, %q, %q, want a,b,c", a, b, c)
	}
}
