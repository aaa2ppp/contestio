package contestio

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func Benchmark_scanInt(b *testing.B) {
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
				fmt.Fscan(br, &res[i])
			}
		}
	})

	b.Run("ReadString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			s, _ := br.ReadString('\n')
			tokens := strings.Fields(s)
			for i, token := range tokens {
				v, _ := strconv.Atoi(token)
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
				v, _ := strconv.Atoi(sc.Text())
				res[i] = v
			}
		}
	})

	b.Run("ScanInts", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			if _, err := ScanInts(br, res); err != nil {
				b.Fatalf("ScanInts: %v", err)
			}
		}
	})

	b.Run("ScanIntsLn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:0]
			b.StartTimer()

			res, err := ScanIntsLn(br, res)
			if err != nil {
				b.Fatalf("ScanInts: %v", err)
			}
		}
	})

	b.Run("ScanInt_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory
			b.StartTimer()

			for j := 0; j+3 < len(res); j += 3 {
				if _, err := ScanInt(br, &res[j], &res[j+1], &res[j+2]); err != nil {
					b.Fatalf("ScanInt: %v", err)
				}
			}
		}
	})

	b.Run("ScanAny_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory
			b.StartTimer()

			for j := 0; j+3 < len(res); j += 3 {
				if _, err := ScanAny(br, &res[j], &res[j+1], &res[j+2]); err != nil {
					b.Fatalf("ScanAny: %v", err)
				}
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
				fmt.Fscan(br, &res[i])
			}
		}
	})

	b.Run("ReadString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := bufio.NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			s, _ := br.ReadString('\n')
			tokens := strings.Fields(s)

			for i, token := range tokens {
				v, _ := strconv.ParseFloat(token, 64)
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
				v, _ := strconv.ParseFloat(sc.Text(), 64)
				res[i] = v
			}
		}
	})

	b.Run("ScanFloats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:N]
			b.StartTimer()

			if _, err := ScanFloats(br, res); err != nil {
				b.Fatalf("ScanFloats: %v", err)
			}
		}
	})

	b.Run("ScanFloatsLn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory[:0]
			b.StartTimer()

			res, err := ScanFloatsLn(br, res[:0])
			if err != nil {
				b.Fatalf("ScanFloatsLn: %v", err)
			}
		}
	})

	b.Run("scanFloat_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory
			b.StartTimer()

			for j := 0; j+3 < len(res); j += 3 {
				if _, err := ScanFloat(br, &res[j], &res[j+1], &res[j+2]); err != nil {
					b.Fatalf("ScanFloat: %v", err)
				}
			}
		}
	})

	b.Run("scanAny_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			br := NewReader(bytes.NewReader(input))
			res := memory
			b.StartTimer()

			for j := 0; j+3 < len(res); j += 3 {
				if _, err := ScanAny(br, &res[j], &res[j+1], &res[j+2]); err != nil {
					b.Fatalf("ScanAny: %v", err)
				}
			}
		}
	})
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

	b.Run("PrintInts", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			if _, err := PrintInts(bw, _lineWO, nums); err != nil {
				b.Fatalf("PrintInts: %v", err)
			}
		}
	})

	b.Run("PrintInt_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			op := WO{Sep: " ", End: " "}
			for j := 0; j+3 < len(nums); j += 3 {
				if _, err := PrintInt(bw, op, nums[j:j+3]...); err != nil {
					b.Fatalf("PrintInt: %v", err)
				}
			}
		}
	})

	b.Run("PrintAny_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			op := WO{Sep: " ", End: " "}
			for j := 0; j+3 < len(nums); j += 3 {
				if _, err := PrintAny(bw, op, &nums[j], &nums[j+1], &nums[j+2]); err != nil {
					b.Fatalf("PrintAny: %v", err)
				}
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

	b.Run("PrintFloats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			PrintFloats(bw, _lineWO, nums)
		}
	})

	b.Run("PrintFloat_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			op := WO{Sep: " ", End: " "}
			for j := 0; j+3 < len(nums); j += 3 {
				if _, err := PrintFloat(bw, op, nums[j:j+3]...); err != nil {
					b.Fatalf("PrintFloat: %v", err)
				}
			}
		}
	})

	b.Run("PrintAny_loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bw := NewWriter(io.Discard)
			b.StartTimer()

			op := WO{Sep: " ", End: " "}
			for j := 0; j+3 < len(nums); j += 3 {
				if _, err := PrintAny(bw, op, &nums[j], &nums[j+1], &nums[j+2]); err != nil {
					b.Fatalf("PrintAny: %v", err)
				}
			}
		}
	})
}
