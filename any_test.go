//go:build any

package contestio

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"reflect"
	"strconv"
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

func Test_printAnyCommon(t *testing.T) {
	tests := []struct {
		name    string
		opts    writeOpts
		args    []any
		wantOut string
		wantErr error
	}{
		{
			name:    "пустой слайс",
			opts:    lineWO,
			args:    []any{},
			wantOut: "\n",
			wantErr: nil,
		},
		{
			name:    "один элемент (строка)",
			opts:    lineWO,
			args:    []any{"hello"},
			wantOut: "hello\n",
			wantErr: nil,
		},
		{
			name:    "несколько строк",
			opts:    lineWO,
			args:    []any{"a", "b", "c"},
			wantOut: "a b c\n",
			wantErr: nil,
		},
		{
			name:    "целые числа",
			opts:    lineWO,
			args:    []any{42, -7, 0},
			wantOut: "42 -7 0\n",
			wantErr: nil,
		},
		{
			name:    "беззнаковые целые",
			opts:    lineWO,
			args:    []any{uint(10), uint8(255), uint64(100500)},
			wantOut: "10 255 100500\n",
			wantErr: nil,
		},
		{
			name:    "числа с плавающей точкой",
			opts:    lineWO,
			args:    []any{3.14, -0.5, 0.0},
			wantOut: "3.14 -0.5 0\n",
			wantErr: nil,
		},
		{
			name:    "смешанные типы (строки, числа)",
			opts:    lineWO,
			args:    []any{"answer", 42, 3.14},
			wantOut: "answer 42 3.14\n",
			wantErr: nil,
		},
		{
			name:    "смешанные типы указатели (строки, числа)",
			opts:    lineWO,
			args:    func() []any { s := "answer"; i := 42; f := 3.14; return []any{&s, &i, &f} }(),
			wantOut: "answer 42 3.14\n",
			wantErr: nil,
		},
		{
			name:    "кастомный разделитель и обрамление",
			opts:    writeOpts{Begin: "[", Sep: ",", End: "]"},
			args:    []any{1, 2, 3},
			wantOut: "[1,2,3]",
			wantErr: nil,
		},
		{
			name:    "пустой Begin и End",
			opts:    writeOpts{Begin: "", Sep: ",", End: ""},
			args:    []any{"x", "y"},
			wantOut: "x,y",
			wantErr: nil,
		},
		{
			name:    "только Sep",
			opts:    writeOpts{Begin: "", Sep: "|", End: ""},
			args:    []any{1, 2, 3},
			wantOut: "1|2|3",
			wantErr: nil,
		},
		{
			name:    "один элемент с Begin и End",
			opts:    writeOpts{Begin: "(", Sep: ",", End: ")"},
			args:    []any{42},
			wantOut: "(42)",
			wantErr: nil,
		},
		{
			name:    "неподдерживаемый тип (bool)",
			opts:    lineWO,
			args:    []any{true},
			wantOut: "",
			wantErr: errors.New("unsupported kind: bool"),
		},
		{
			name:    "неподдерживаемый тип (структура)",
			opts:    lineWO,
			args:    []any{struct{}{}},
			wantOut: "",
			wantErr: errors.New("unsupported kind: struct"),
		},
		{
			name:    "неподдерживаемый тип в середине",
			opts:    lineWO,
			args:    []any{1, struct{}{}, 2},
			wantOut: "1 ", // первый элемент и разделитель уже выведены
			wantErr: errors.New("unsupported kind: struct"),
		},
		{
			name:    "срез (неподдерживаемый)",
			opts:    lineWO,
			args:    []any{[]int{1, 2}},
			wantOut: "",
			wantErr: errors.New("unsupported kind: slice"),
		},
		{
			name:    "map (неподдерживаемый)",
			opts:    lineWO,
			args:    []any{map[string]int{}},
			wantOut: "",
			wantErr: errors.New("unsupported kind: map"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)

			n, err := printAnyCommon(w, tt.opts, tt.args...)

			// Сбрасываем буфер, чтобы все данные попали в buf
			_ = w.Flush()

			// Проверяем ошибку
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ожидалась ошибка %q, но ошибки нет", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("ошибка = %q, ожидалась %q", err, tt.wantErr)
				}
				// Если ожидалась ошибка, то проверяем, что количество обработанных элементов
				// соответствует позиции, на которой произошёл сбой.
				// Для неподдерживаемого типа в середине n должно быть равно индексу этого элемента,
				// потому что предыдущие элементы были успешно записаны.
				if tt.args != nil && tt.wantErr != nil && len(tt.args) > 0 {
					// Находим первый неподдерживаемый элемент
					for i, v := range tt.args {
						k := reflect.TypeOf(v).Kind()
						if k == reflect.String {
							continue
						}
						if uint(k) >= uint(len(appendAnyTab)) || appendAnyTab[k] == nil {
							if n != i {
								t.Errorf("количество обработанных элементов = %d, ожидалось %d (индекс первого неподдерживаемого элемента)", n, i)
							}
							break
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("неожиданная ошибка: %v", err)
				}
				// Проверяем вывод
				if got := buf.String(); got != tt.wantOut {
					t.Errorf("вывод = %q, ожидалось %q", got, tt.wantOut)
				}
				// Проверяем, что количество обработанных элементов равно длине слайса
				if n != len(tt.args) {
					t.Errorf("возвращённое количество элементов = %d, ожидалось %d", n, len(tt.args))
				}
			}
		})
	}
}

func Benchmark_scanAny(b *testing.B) {
	b.Run("Int", func(b *testing.B) {
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
			for j := 0; j+3 < len(res); j += 3 {
				scanAnyCommon(r, false, &res[j], &res[j+1], &res[j+2])
			}
			switch len(res) % 3 {
			case 1:
				scanAnyCommon(r, false, &res[len(res)-1])
			case 2:
				scanAnyCommon(r, false, &res[len(res)-2], &res[len(res)-1])
			}
		}
	})
	b.Run("Float", func(b *testing.B) {
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
			for j := 0; j+3 < len(res); j += 3 {
				scanAnyCommon(r, false, &res[j], &res[j+1], &res[j+2])
			}
			switch len(res) % 3 {
			case 1:
				scanAnyCommon(r, false, &res[len(res)-1])
			case 2:
				scanAnyCommon(r, false, &res[len(res)-2], &res[len(res)-1])
			}
		}
	})
	b.Run("Word", func(b *testing.B) {
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
			for j := 0; j+3 < len(res); j += 3 {
				scanAnyCommon(r, false, &res[j], &res[j+1], &res[j+2])
			}
			switch len(res) % 3 {
			case 1:
				scanAnyCommon(r, false, &res[len(res)-1])
			case 2:
				scanAnyCommon(r, false, &res[len(res)-2], &res[len(res)-1])
			}
		}
	})
}

func Benchmark_printAny(b *testing.B) {
	N := 1 << 20
	b.Run("Int", func(b *testing.B) {
		rand := rand.New(rand.NewSource(1))
		nums := generateInts[int](rand, N)
		w := NewWriter(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j+3 < len(nums); j += 3 {
				PrintAny(w, lineWO, &nums[j], &nums[j+1], &nums[j+2])
			}
			switch len(nums) % 3 {
			case 1:
				PrintAny(w, lineWO, &nums[len(nums)-1])
			case 2:
				PrintAny(w, lineWO, &nums[len(nums)-2], &nums[len(nums)-1])
			}
		}
	})
	b.Run("Float", func(b *testing.B) {
		rand := rand.New(rand.NewSource(1))
		nums := generateFloats[float64](rand, N)
		w := NewWriter(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j+3 < len(nums); j += 3 {
				PrintAny(w, lineWO, &nums[j], &nums[j+1], &nums[j+2])
			}
			switch len(nums) % 3 {
			case 1:
				PrintAny(w, lineWO, &nums[len(nums)-1])
			case 2:
				PrintAny(w, lineWO, &nums[len(nums)-2], &nums[len(nums)-1])
			}
		}
	})
	b.Run("String", func(b *testing.B) {
		rand := rand.New(rand.NewSource(1))
		nums := generateFloats[float64](rand, N)
		words := make([]string, 0, len(nums))
		for _, v := range nums {
			words = append(words, strconv.FormatFloat(v, 'g', -1, 64))
		}
		w := NewWriter(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j+3 < len(nums); j += 3 {
				PrintAny(w, lineWO, &nums[j], &nums[j+1], &nums[j+2])
			}
			switch len(nums) % 3 {
			case 1:
				PrintAny(w, lineWO, &nums[len(nums)-1])
			case 2:
				PrintAny(w, lineWO, &nums[len(nums)-2], &nums[len(nums)-1])
			}
		}
	})
}
