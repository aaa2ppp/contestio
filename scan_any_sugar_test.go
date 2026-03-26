//go:build sugar

package contestio

import (
	"bytes"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func Test_scanAnyCommon(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		stopAtEol  bool
		args       []any
		wantN      int
		wantErr    bool
		wantValues []any
	}{
		{
			name:       "two ints",
			input:      "123 456",
			stopAtEol:  false,
			args:       []any{new(int), new(int)},
			wantN:      2,
			wantValues: []any{123, 456},
		},
		{
			name:       "int and float",
			input:      "10 3.14",
			stopAtEol:  false,
			args:       []any{new(int), new(float64)},
			wantN:      2,
			wantValues: []any{10, 3.14},
		},
		{
			name:       "int and string",
			input:      "42 hello",
			stopAtEol:  false,
			args:       []any{new(int), new(string)},
			wantN:      2,
			wantValues: []any{42, "hello"},
		},
		{
			name:      "EOF before enough tokens",
			input:     "123",
			stopAtEol: false,
			args:      []any{new(int), new(int)},
			wantN:     1,
			wantErr:   true, // io.ErrUnexpectedEOF
		},
		{
			name:      "empty input",
			input:     "",
			stopAtEol: false,
			args:      []any{new(int)},
			wantN:     0,
			wantErr:   true, // io.EOF
		},
		{
			name:       "stopAtEol stops at newline (one arg)",
			input:      "123\n456",
			stopAtEol:  true,
			args:       []any{new(int)},
			wantN:      1,
			wantValues: []any{123},
		},
		{
			name:      "stopAtEol with two args, newline after first",
			input:     "123\n456",
			stopAtEol: true,
			args:      []any{new(int), new(int)},
			wantN:     1,
			wantErr:   true, // skipSpace вернёт EOL при попытке читать второй аргумент
		},
		{
			name:      "non-pointer argument",
			input:     "123",
			stopAtEol: false,
			args:      []any{int(0)},
			wantN:     0,
			wantErr:   true,
		},
		{
			name:      "unsupported kind (bool)",
			input:     "true",
			stopAtEol: false,
			args:      []any{new(bool)},
			wantN:     0,
			wantErr:   true,
		},
		// nil -> always panic
		// {
		// 	name:    "untyped nil",
		// 	input:   "123",
		// 	args:    []any{nil},
		// 	wantN:   0,
		// 	wantErr: true,
		// },
		// {
		// 	name:    "typed nil pointer",
		// 	input:   "123",
		// 	args:    []any{(*int)(nil)},
		// 	wantN:   0,
		// 	wantErr: true,
		// },
		{
			name:    "mixed: valid and nil",
			args:    []any{new(int), (*float64)(nil)},
			wantErr: true,
		},

		{
			name:       "mixed types with extra spaces",
			input:      "   -5  \t 2.718  \n  word  ",
			stopAtEol:  false,
			args:       []any{new(int), new(float64), new(string)},
			wantN:      3,
			wantValues: []any{-5, 2.718, "word"},
		},
		{
			name:       "MyInt/Float/String",
			input:      "   -5  \t 2.718  \n  word  ",
			stopAtEol:  false,
			args:       []any{new(MyInt), new(MyFloat), new(MyString)},
			wantN:      3,
			wantValues: []any{MyInt(-5), MyFloat(2.718), MyString("word")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			n, err := scanAnyCommon(br, tt.stopAtEol, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if n != tt.wantN {
				t.Errorf("n = %v, want %v", n, tt.wantN)
			}
			if err == nil && tt.wantValues != nil {
				for i, want := range tt.wantValues {
					got := reflect.ValueOf(tt.args[i]).Elem().Interface()
					if !reflect.DeepEqual(got, want) {
						t.Errorf("arg %d = %v, want %v", i, got, want)
					}
				}
			}
		})
	}
}

func Test_scanAnyLnCommon(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		args       []any
		wantN      int
		wantErr    bool
		wantValues []any
	}{
		{
			name:       "full line",
			input:      "10 20\n",
			args:       []any{new(int), new(int)},
			wantN:      2,
			wantValues: []any{10, 20},
		},
		{
			name:    "extra tokens after line",
			input:   "10 20 30\n",
			args:    []any{new(int), new(int)},
			wantN:   2,
			wantErr: true, // ожидается EOL, но есть ещё токен
		},
		{
			name:       "EOF after line (no newline)",
			input:      "10 20",
			args:       []any{new(int), new(int)},
			wantN:      2,
			wantValues: []any{10, 20},
		},
		{
			name:    "incomplete line",
			input:   "10",
			args:    []any{new(int), new(int)},
			wantN:   1,
			wantErr: true, // не хватает данных
		},
		{
			name:    "empty line",
			input:   "\n",
			args:    []any{new(int)},
			wantN:   0,
			wantErr: true, // нет токенов
		},
		{
			name:       "line with trailing spaces",
			input:      "  42  \t  3.14  \n",
			args:       []any{new(int), new(float64)},
			wantN:      2,
			wantValues: []any{42, 3.14},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			n, err := scanAnyLnCommon(br, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if n != tt.wantN {
				t.Errorf("n = %v, want %v", n, tt.wantN)
			}
			if err == nil && tt.wantValues != nil {
				for i, want := range tt.wantValues {
					got := reflect.ValueOf(tt.args[i]).Elem().Interface()
					if !reflect.DeepEqual(got, want) {
						t.Errorf("arg %d = %v, want %v", i, got, want)
					}
				}
			}
		})
	}
}

func Benchmark_scanAny(b *testing.B) {
	b.Run("patseInt_loop", func(b *testing.B) {
		N := 1 << 20
		rand := rand.New(rand.NewSource(1))
		nums := generateInts[int](rand, N)
		input := makeIntsInput(rand, nums, 5)
		memory := make([]int, N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			r := NewReader(bytes.NewReader(input))
			res := memory
			b.StartTimer()
			for j := 0; j+3 < len(nums); j += 3 {
				scanAnyCommon(r, false, &res[j], &res[j+1], &res[j+2])
			}
		}
	})
	b.Run("patseFloat_loop", func(b *testing.B) {
		N := 1 << 20
		rand := rand.New(rand.NewSource(1))
		nums := generateFloats[float64](rand, N)
		input := makeFloatsInput(rand, nums, 5)
		memory := make([]float64, N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			r := NewReader(bytes.NewReader(input))
			res := memory
			b.StartTimer()
			for j := 0; j+3 < len(nums); j += 3 {
				scanAnyCommon(r, false, &res[j], &res[j+1], &res[j+2])
			}
		}
	})
	b.Run("patseWord_loop", func(b *testing.B) {
		N := 1 << 20
		rand := rand.New(rand.NewSource(1))
		nums := generateFloats[float64](rand, N)
		input := makeFloatsInput(rand, nums, 5)
		memory := make([]string, N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			r := NewReader(bytes.NewReader(input))
			res := memory
			b.StartTimer()
			for j := 0; j+3 < len(nums); j += 3 {
				scanAnyCommon(r, false, &res[j], &res[j+1], &res[j+2])
			}
		}
	})
}
