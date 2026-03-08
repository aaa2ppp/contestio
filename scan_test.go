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
			io.EOF,
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
			io.EOF,
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
			io.EOF,
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
			io.EOF,
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
			br := bufio.NewReader(bytes.NewReader(input))
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
			br := bufio.NewReader(bytes.NewReader(input))
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
			br := bufio.NewReader(bytes.NewReader(input))
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
			br := bufio.NewReader(bytes.NewReader(input))
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
			br := bufio.NewReader(bytes.NewReader(input))
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
			br := bufio.NewReader(bytes.NewReader(input))
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
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			n, err := scanSlice(br, parseIntStd, res)

			if n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})
	b.Run("parseIntBase", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			n, err := scanSlice(br, parseInt, res)

			if n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})
	b.Run("parseIntFast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			n, err := scanSlice(br, parseIntFast, res)

			if n != N || err != nil {
				b.Errorf("unexpected result (%v, %v), want (%d, nil)", n, err, N)
				return
			}
		}
	})
}
