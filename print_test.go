package contestio

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func Test_printSlice(t *testing.T) {
	sliceWO := writeOpts{Begin: "[]int{", Sep: ", ", End: "}"}

	tests := []struct {
		name    string
		bufSize int
		op      writeOpts
		a       []int
		wantOut string
		wantErr bool
	}{
		{
			"simple",
			16,
			sliceWO,
			[]int{1, 2, 3, 4, 5, 6},
			"[]int{1, 2, 3, 4, 5, 6}",
			false,
		},
		{
			"one",
			16,
			sliceWO,
			[]int{1},
			"[]int{1}",
			false,
		},
		{
			"empty",
			16,
			sliceWO,
			[]int{},
			"[]int{}",
			false,
		},
		{
			"nil",
			16,
			sliceWO,
			nil,
			"[]int{}",
			false,
		},
		{
			"over buffer",
			16,
			sliceWO,
			make([]int, 100),
			"[]int{" + strings.Repeat("0, ", 100)[:298] + "}",
			false,
		},
		{
			"zero opts",
			16,
			writeOpts{},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			"1234567890",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			bw := NewWriterSize(&out, tt.bufSize)

			gotN, gotErr := printSliceCommon(bw, tt.op, appendInt[int], tt.a)
			bw.Flush()

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("printSlice() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("printSlice() succeeded unexpectedly")
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("printSlice() out = %v, want %v", gotOut, tt.wantOut)
			}
			if gotN != len(tt.a) {
				t.Errorf("printSlice() = %v, want %v", gotN, len(tt.a))
			}
		})
	}
}

func Benchmark_printInt(b *testing.B) {
	N := 1 << 20
	rand := rand.New(rand.NewSource(1))
	nums := generateInts[int](rand, N)

	b.Run("fmt.Fprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := bufio.NewWriter(io.Discard)
			b.StartTimer()

			for _, v := range nums {
				fmt.Fprintf(bw, "%d ", v)
			}
		}
	})

	b.Run("AppendInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := bufio.NewWriter(io.Discard)
			b.StartTimer()

			for _, v := range nums {
				bw.Write(strconv.AppendInt(bw.AvailableBuffer(), int64(v), 10))
				bw.WriteByte(' ')
			}
		}
	})

	b.Run("AppendInt_scr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := bufio.NewWriter(io.Discard)
			b.StartTimer()

			var scratch [32]byte
			var buf []byte
			for _, v := range nums {
				if bw.Available() < len(scratch) {
					buf = scratch[:0]
				} else {
					buf = bw.AvailableBuffer()
				}
				buf = strconv.AppendInt(buf, int64(v), 10)
				bw.Write(buf)
				bw.WriteByte(' ')
			}
		}
	})

	b.Run("printSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			printSlice(bw, lineWO, appendInt, nums)
		}
	})

	b.Run("printVals_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			for _, v := range nums {
				printVals(bw, lineWO, appendInt, v)
			}
		}
	})
}

func Benchmark_printFloat(b *testing.B) {
	N := 1 << 20
	rand := rand.New(rand.NewSource(2))
	nums := generateFloats[float64](rand, N)

	b.Run("fmt.Fprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := bufio.NewWriter(io.Discard)
			b.StartTimer()

			for _, v := range nums {
				fmt.Fprintf(bw, "%g ", v)
			}
		}
	})

	b.Run("AppendFloat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := bufio.NewWriter(io.Discard)
			b.StartTimer()

			for _, v := range nums {
				bw.Write(strconv.AppendFloat(bw.AvailableBuffer(), v, 'g', -1, 64))
				bw.WriteByte(' ')
			}
		}
	})

	b.Run("AppendFloat_scr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := bufio.NewWriter(io.Discard)
			b.StartTimer()

			var scratch [32]byte
			var buf []byte
			for _, v := range nums {
				if bw.Available() < len(scratch) {
					buf = scratch[:0]
				} else {
					buf = bw.AvailableBuffer()
				}
				buf = strconv.AppendFloat(buf, v, 'g', -1, 64)
				bw.Write(buf)
				bw.WriteByte(' ')
			}
		}
	})

	b.Run("printSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			printSlice(bw, lineWO, appendFloat, nums)
		}
	})

	b.Run("printVals_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			for j := 0; j+3 < len(nums); j += 3 {
				printVals(bw, lineWO, appendFloat, nums[j:j+3]...)
			}
		}
	})
}
