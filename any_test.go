//go:build any

package contestio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

func TestScanAny(t *testing.T) {
	panicErr := errors.New("panic")

	tests := []struct {
		name       string
		input      string
		args       []any
		wantN      int
		wantErr    bool
		wantPanic  bool
		wantValues []any
	}{
		{
			name:       "two ints",
			input:      "123 456",
			args:       []any{new(int), new(int)},
			wantN:      2,
			wantValues: []any{123, 456},
		},
		{
			name:       "int and float",
			input:      "10 3.14",
			args:       []any{new(int), new(float64)},
			wantN:      2,
			wantValues: []any{10, 3.14},
		},
		{
			name:       "int and string",
			input:      "42 hello",
			args:       []any{new(int), new(string)},
			wantN:      2,
			wantValues: []any{42, "hello"},
		},
		{
			name:    "EOF before enough tokens",
			input:   "123",
			args:    []any{new(int), new(int)},
			wantN:   1,
			wantErr: true, // io.ErrUnexpectedEOF
		},
		{
			name:    "empty input",
			input:   "",
			args:    []any{new(int)},
			wantN:   0,
			wantErr: true, // io.EOF
		},
		{
			name:    "non-pointer argument",
			input:   "123",
			args:    []any{int(0)},
			wantN:   0,
			wantErr: true,
		},
		{
			name:    "unsupported kind (bool)",
			input:   "true",
			args:    []any{new(bool)},
			wantN:   0,
			wantErr: true,
		},
		{
			name:      "untyped nil",
			input:     "123",
			args:      []any{nil},
			wantN:     0,
			wantErr:   true,
			wantPanic: true,
		},
		{
			name:      "typed nil pointer",
			input:     "123",
			args:      []any{(*int)(nil)},
			wantN:     0,
			wantErr:   true,
			wantPanic: true,
		},
		{
			name:    "mixed: valid and nil",
			args:    []any{new(int), (*float64)(nil)},
			wantErr: true,
		},

		{
			name:       "mixed types with extra spaces",
			input:      "   -5  \t 2.718  \n  word  ",
			args:       []any{new(int), new(float64), new(string)},
			wantN:      3,
			wantValues: []any{-5, 2.718, "word"},
		},
		{
			name:       "MyInt/MyFloat/MyString",
			input:      "   -5  \t 2.718  \n  word  ",
			args:       []any{new(MyInt), new(MyFloat), new(MyString)},
			wantN:      3,
			wantValues: []any{MyInt(-5), MyFloat(2.718), MyString("word")},
		},
	}

	catchPanic := func(fn func(*Reader, ...any) (int, error), br *Reader, a ...any) (_ int, err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("%w: %v", panicErr, p)
			}
		}()
		return fn(br, a...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewReader(strings.NewReader(tt.input))
			n, err := catchPanic(ScanAny, br, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if errors.Is(err, panicErr) != tt.wantPanic {
				t.Errorf("panic = %v, wantPanic %v", err, tt.wantErr)
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

func TestScanAnyLn(t *testing.T) {
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
			n, err := ScanAnyLn(br, tt.args...)
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

func TestPrintAny(t *testing.T) {
	panicErr := errors.New("panic")

	tests := []struct {
		name      string
		opts      writeOpts
		args      []any
		wantN     int
		wantErr   bool
		wantPanic bool
		wantOut   string
	}{
		{
			name:    "empty slice",
			opts:    lineWO,
			args:    []any{},
			wantN:   0,
			wantOut: "\n",
		},
		{
			name:    "single string",
			opts:    lineWO,
			args:    []any{"hello"},
			wantN:   1,
			wantOut: "hello\n",
		},
		{
			name:    "multiple strings",
			opts:    lineWO,
			args:    []any{"a", "b", "c"},
			wantN:   3,
			wantOut: "a b c\n",
		},
		{
			name:    "integers",
			opts:    lineWO,
			args:    []any{42, -7, 0},
			wantN:   3,
			wantOut: "42 -7 0\n",
		},
		{
			name:    "unsigned integers",
			opts:    lineWO,
			args:    []any{uint(10), uint8(255), uint64(100500)},
			wantN:   3,
			wantOut: "10 255 100500\n",
		},
		{
			name:    "floats",
			opts:    lineWO,
			args:    []any{3.14, -0.5, 0.0},
			wantN:   3,
			wantOut: "3.14 -0.5 0\n",
		},
		{
			name:    "mixed types",
			opts:    lineWO,
			args:    []any{"answer", 42, 3.14},
			wantN:   3,
			wantOut: "answer 42 3.14\n",
		},
		{
			name:    "pointers to mixed types",
			opts:    lineWO,
			args:    func() []any { s := "answer"; i := 42; f := 3.14; return []any{&s, &i, &f} }(),
			wantN:   3,
			wantOut: "answer 42 3.14\n",
		},
		{
			name:    "custom separator and brackets",
			opts:    writeOpts{Begin: "[", Sep: ",", End: "]"},
			args:    []any{1, 2, 3},
			wantN:   3,
			wantOut: "[1,2,3]",
		},
		{
			name:    "empty begin and end",
			opts:    writeOpts{Begin: "", Sep: ",", End: ""},
			args:    []any{"x", "y"},
			wantN:   2,
			wantOut: "x,y",
		},
		{
			name:    "only separator",
			opts:    writeOpts{Begin: "", Sep: "|", End: ""},
			args:    []any{1, 2, 3},
			wantN:   3,
			wantOut: "1|2|3",
		},
		{
			name:    "single element with brackets",
			opts:    writeOpts{Begin: "(", Sep: ",", End: ")"},
			args:    []any{42},
			wantN:   1,
			wantOut: "(42)",
		},
		{
			name:    "unsupported type bool",
			opts:    lineWO,
			args:    []any{true},
			wantN:   0,
			wantErr: true,
		},
		{
			name:    "unsupported type struct",
			opts:    lineWO,
			args:    []any{struct{}{}},
			wantN:   0,
			wantErr: true,
		},
		{
			name:    "unsupported type in the middle",
			opts:    lineWO,
			args:    []any{1, struct{}{}, 2},
			wantN:   1,
			wantErr: true,
		},
		{
			name:    "unsupported type slice",
			opts:    lineWO,
			args:    []any{[]int{1, 2}},
			wantN:   0,
			wantErr: true,
		},
		{
			name:    "unsupported type map",
			opts:    lineWO,
			args:    []any{map[string]int{}},
			wantN:   0,
			wantErr: true,
		},
		{
			name:      "untyped nil",
			opts:      lineWO,
			args:      []any{nil},
			wantN:     0,
			wantErr:   true,
			wantPanic: true,
		},
		{
			name:      "typed nil pointer",
			opts:      lineWO,
			args:      []any{(*int)(nil)},
			wantN:     0,
			wantErr:   true,
			wantPanic: true,
		},
	}

	catchPanic := func(fn func(*Writer, writeOpts, ...any) (int, error), w *Writer, opts writeOpts, a ...any) (_ int, err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("%w: %v", panicErr, p)
			}
		}()
		return fn(w, opts, a...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)

			n, err := catchPanic(PrintAny, w, tt.opts, tt.args...)
			_ = w.Flush()

			// TODO: переписать, чтобы было совместимо с флагом must
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if errors.Is(err, panicErr) != tt.wantPanic {
				t.Errorf("panic = %v, wantPanic %v", err, tt.wantPanic)
			}

			if n != tt.wantN {
				t.Errorf("n = %v, want %v", n, tt.wantN)
			}
			if err == nil {
				if got := buf.String(); got != tt.wantOut {
					t.Errorf("output = %q, want %q", got, tt.wantOut)
				}
			}
		})
	}
}

const anyBulkN = 1 << 10 // 1K

func test_scanPrintAnyInt_bulk[T Int](t *testing.T) {
	N := anyBulkN
	rand := rand.New(rand.NewSource(1))
	nums := generateInts[T](rand, N)
	input := makeIntsInput(nil, nums, 1)

	t.Run("scan", func(t *testing.T) {
		res := make([]T, len(nums))
		br := NewReader(bytes.NewReader(input))
		for i := range res {
			ScanAny(br, &res[i])
		}
		if !reflect.DeepEqual(res, nums) {
			t.Errorf("scanAnyCommon for %T fail", res[0])
		}
	})

	t.Run("print", func(t *testing.T) {
		var out bytes.Buffer
		bw := NewWriter(&out)
		op := WO{End: " "}
		for i := range nums {
			PrintAny(bw, op, &nums[i])
		}
		bw.Flush()
		wantOut := bytes.TrimSpace(input)
		gotOut := bytes.TrimSpace(out.Bytes())
		if !reflect.DeepEqual(gotOut, wantOut) {
			t.Errorf("printAnyCommon for %T fail", nums[0])
		}
	})
}

func test_scanPrintAnyFloat_bulk[T Float](t *testing.T) {
	N := anyBulkN
	rand := rand.New(rand.NewSource(1))
	nums := generateFloats[T](rand, N)
	input := makeFloatsInput(nil, nums, 1)

	t.Run("sacn", func(t *testing.T) {
		res := make([]T, len(nums))
		br := NewReader(bytes.NewReader(input))
		for i := range res {
			ScanAny(br, &res[i])
		}
		if !reflect.DeepEqual(res, nums) {
			t.Errorf("scanAnyCommon for %T fail", res[0])
		}
	})

	t.Run("print", func(t *testing.T) {
		var out bytes.Buffer
		bw := NewWriter(&out)
		op := WO{End: " "}
		for i := range nums {
			PrintAny(bw, op, &nums[i])
		}
		bw.Flush()
		wantOut := bytes.TrimSpace(input)
		gotOut := bytes.TrimSpace(out.Bytes())
		if !reflect.DeepEqual(gotOut, wantOut) {
			t.Errorf("printAnyCommon for %T fail", nums[0])
		}
	})
}

func test_scanPrintAnyWord_bulk[T ~string](t *testing.T) {
	N := anyBulkN
	rand := rand.New(rand.NewSource(1))
	nums := generateFloats[float64](rand, N)
	input := makeFloatsInput(nil, nums, 1)
	words := strings.Fields(unsafeString(input))

	t.Run("scan", func(t *testing.T) {
		res := make([]T, len(nums))
		br := NewReader(bytes.NewReader(input))
		for i := range res {
			ScanAny(br, &res[i])
		}
		resAsStrings := *(*[]string)(unsafe.Pointer(&res))
		if !reflect.DeepEqual(resAsStrings, words) {
			t.Errorf("scanAnyCommon for %T fail", res[0])
		}
	})

	t.Run("print", func(t *testing.T) {
		var out bytes.Buffer
		bw := NewWriter(&out)
		op := WO{End: " "}
		for i := range words {
			PrintAny(bw, op, &words[i])
		}
		bw.Flush()
		wantOut := bytes.TrimSpace(input)
		gotOut := bytes.TrimSpace(out.Bytes())
		if !reflect.DeepEqual(gotOut, wantOut) {
			t.Errorf("printAnyCommon for %T fail", nums[0])
		}
	})
}

func Test_scanPrintAny_bulk(t *testing.T) {
	t.Run("int8", test_scanPrintAnyInt_bulk[int8])
	t.Run("int16", test_scanPrintAnyInt_bulk[int16])
	t.Run("int32", test_scanPrintAnyInt_bulk[int32])
	t.Run("int64", test_scanPrintAnyInt_bulk[int64])

	t.Run("uint8", test_scanPrintAnyInt_bulk[uint8])
	t.Run("uint16", test_scanPrintAnyInt_bulk[uint16])
	t.Run("uint32", test_scanPrintAnyInt_bulk[uint32])
	t.Run("uint64", test_scanPrintAnyInt_bulk[uint64])

	t.Run("float32", test_scanPrintAnyFloat_bulk[float32])
	t.Run("float64", test_scanPrintAnyFloat_bulk[float64])

	t.Run("string", test_scanPrintAnyWord_bulk[string])

	t.Run("MyInt", test_scanPrintAnyInt_bulk[MyInt])
	t.Run("MyFloat", test_scanPrintAnyFloat_bulk[MyFloat])
	t.Run("MyString", test_scanPrintAnyWord_bulk[MyString])
}

func BenchmarkScanAny(b *testing.B) {
	N := 1 << 20
	b.Run("Int", func(b *testing.B) {
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
				if n, err := ScanAny(r, &res[j], &res[j+1], &res[j+2]); err != nil {
					b.Fatalf("ScanAny: %d, %v", n, err)
				}
			}
		}
	})
	b.Run("Float", func(b *testing.B) {
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
				if n, err := ScanAny(r, &res[j], &res[j+1], &res[j+2]); err != nil {
					b.Fatalf("ScanAny: %d, %v", n, err)
				}
			}
		}
	})
	b.Run("Word", func(b *testing.B) {
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
				if n, err := ScanAny(r, &res[j], &res[j+1], &res[j+2]); err != nil {
					b.Fatalf("ScanAny: %d, %v", n, err)
				}
			}
		}
	})
}

func BenchmarkPrintAnyVal(b *testing.B) {
	N := anyBulkN

	b.Run("Int", func(b *testing.B) {
		rand := rand.New(rand.NewSource(1))
		nums := generateInts[int](rand, N)
		w := NewWriter(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var v1, v2, v3 int
			for j := 0; j+3 < len(nums); j += 3 {
				v1, v2, v3 = nums[j], nums[j+1], nums[j+2]
				if n, err := PrintAny(w, lineWO, v1, v2, v3); err != nil {
					b.Fatalf("PrintAny: %d, %v", n, err)
				}
			}
		}
	})

	b.Run("Float", func(b *testing.B) {
		rand := rand.New(rand.NewSource(1))
		nums := generateFloats[float64](rand, N)
		w := NewWriter(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var v1, v2, v3 float64
			for j := 0; j+3 < len(nums); j += 3 {
				v1, v2, v3 = nums[j], nums[j+1], nums[j+2]
				if n, err := PrintAny(w, lineWO, v1, v2, v3); err != nil {
					b.Fatalf("PrintAny: %d, %v", n, err)
				}
			}
		}
	})

	b.Run("String", func(b *testing.B) {
		rand := rand.New(rand.NewSource(1))
		nums := generateFloats[float64](rand, N)
		words := strings.Fields(unsafeString(makeFloatsInput(rand, nums, 1)))
		w := NewWriter(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var v1, v2, v3 string
			for j := 0; j+3 < len(words); j += 3 {
				v1, v2, v3 = words[j], words[j+1], words[j+2]
				if n, err := PrintAny(w, lineWO, v1, v2, v3); err != nil {
					b.Fatalf("PrintAny: %d, %v", n, err)
				}
			}
		}
	})
}

func BenchmarkPrintAnyPtr(b *testing.B) {
	N := anyBulkN

	b.Run("Int", func(b *testing.B) {
		rand := rand.New(rand.NewSource(1))
		nums := generateInts[int](rand, N)
		w := NewWriter(io.Discard)
		b.ResetTimer()
		var v1, v2, v3 int // вне цикла, т.к. всегда убегают на кучу
		for i := 0; i < b.N; i++ {
			for j := 0; j+3 < len(nums); j += 3 {
				v1, v2, v3 = nums[j], nums[j+1], nums[j+2]
				if n, err := PrintAny(w, lineWO, &v1, &v2, &v3); err != nil {
					b.Fatalf("PrintAny: %d, %v", n, err)
				}
			}
		}
	})

	b.Run("Float", func(b *testing.B) {
		rand := rand.New(rand.NewSource(1))
		nums := generateFloats[float64](rand, N)
		w := NewWriter(io.Discard)
		b.ResetTimer()
		var v1, v2, v3 float64 // вне цикла, т.к. всегда убегают на кучу
		for i := 0; i < b.N; i++ {
			for j := 0; j+3 < len(nums); j += 3 {
				v1, v2, v3 = nums[j], nums[j+1], nums[j+2]
				if n, err := PrintAny(w, lineWO, &v1, &v2, &v3); err != nil {
					b.Fatalf("PrintAny: %d, %v", n, err)
				}
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
		var v1, v2, v3 string // вне цикла, т.к. всегда убегают на кучу
		for i := 0; i < b.N; i++ {
			for j := 0; j+3 < len(words); j += 3 {
				v1, v2, v3 = words[j], words[j+1], words[j+2]
				if n, err := PrintAny(w, lineWO, &v1, &v2, &v3); err != nil {
					b.Fatalf("PrintAny: %d, %v", n, err)
				}
			}
		}
	})
}
