package contestio

import (
	"hash/fnv"
	"math/rand"
	"strconv"
	"strings"
	"unsafe"
)

func generateInts[T Int](rand *rand.Rand, n int) []T {
	size := int(unsafe.Sizeof(T(0)))
	nums := make([]T, n)
	for i := range nums {
		bits := (rand.Intn(size) + 1) << 3
		mask := ^(^uint64(0) << bits) | (1 << (size<<3 - 1))
		num := T(rand.Uint64() & mask)
		nums[i] = num
	}
	return nums
}

func generateFloats[T Float](rand *rand.Rand, n int) []T {
	nums := make([]T, n)
	for i := range nums {
		nums[i] = T(rand.Float64())*(1<<21) - (1 << 20) // |x| <= 10^6
	}
	return nums
}

func makeAppendSpace(rand *rand.Rand, n int) func([]byte) []byte {
	if n <= 1 {
		return func(b []byte) []byte { return append(b, ' ') }
	}
	return func(b []byte) []byte { return append(b, strings.Repeat(" ", rand.Intn(n)+1)...) }
}

func makeIntsInput[T Int](rand *rand.Rand, nums []T, maxSpace int) []byte {
	appendSpace := makeAppendSpace(rand, maxSpace)
	var input []byte
	for _, v := range nums {
		if ^T(0) < 0 {
			input = strconv.AppendInt(input, int64(v), 10)
		} else {
			input = strconv.AppendUint(input, uint64(v), 10)
		}
		input = appendSpace(input)
	}
	input = append(input, '\n')
	return input
}

func makeFloatsInput[T Float](rand *rand.Rand, nums []T, maxSpace int) []byte {
	appendSpace := makeAppendSpace(rand, maxSpace)
	bitSize := int(unsafe.Sizeof(T(0))) << 3
	var input []byte
	for _, v := range nums {
		input = strconv.AppendFloat(input, float64(v), 'g', -1, bitSize)
		input = appendSpace(input)
	}
	input = append(input, '\n')
	return input
}

func seedFromBytes(data []byte) int64 {
	h := fnv.New64a()
	h.Write(data)
	return int64(h.Sum64())
}
