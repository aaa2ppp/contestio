package contestio

import (
	"math"
	"math/bits"
	"strconv"
	"unsafe"

	"golang.org/x/sys/cpu"
)

// tokenToDigits преобразуем 8 символов за раз. (!) len(token) must be >=8
func tokenToDigits(token []byte) uint64 {
	digits := *(*uint64)(unsafe.Pointer(&token[0]))
	digits ^= 0x3030303030303030
	return digits
}

// checkDigits проверяем 8 цифр за раз
func checkDigits(digits uint64) bool {
	const maskF0 = 0xF0F0F0F0F0F0F0F0
	if digits&maskF0 != 0 || // 0123456789:;<=>?.
		(digits+0x0606060606060606)&maskF0 != 0 { // 0123456789
		return false
	}
	return true
}

// parseDigits обрабатываем по 8 цифр за раз (только три умножения)
func parseDigits(digits uint64) uint64 {
	if !cpu.IsBigEndian {
		digits = bits.ReverseBytes64(digits)
	}
	tens := ((digits >> 8) & 0x00FF00FF00FF00FF) * 10
	pairs := tens + (digits & 0x00FF00FF00FF00FF)

	p_high := ((pairs >> 16) & 0x0000FFFF0000FFFF) * 100
	quads := p_high + (pairs & 0x0000FFFF0000FFFF)

	hi := quads >> 32
	lo := quads & 0xFFFFFFFF
	return hi*10000 + lo
}

func parseIntSwar[T Int](token []byte) (T, error) {
	var unsigned = ^T(0) >= 0 // true для unsigned T

	orig := token
	if len(orig) == 0 {
		return 0, &IntError{string(orig), strconv.ErrSyntax}
	}

	// handle sign
	if orig[0] == '-' || orig[0] == '+' {
		if unsigned {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		token = token[1:]
		if len(token) == 0 {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
	}

	// trim leading zeros
	for len(token) > 0 && token[0] == '0' {
		token = token[1:]
	}

	n := len(token)
	if n == 0 {
		return 0, nil // was "0...0"
	}
	if n > 20 {
		// гарантированное переполнение
		return 0, &IntError{string(orig), strconv.ErrRange}
	}

	var twentiethDigit byte
	if n == 20 {
		twentiethDigit = token[19] - '0' // запоминаем 20-ю цифру
		token = token[:19]               // прячем разряд в котором можем переполнится
	}

	var u64 uint64
	var i int

	// быстро парсим ровно 8 или 16 первых цифр (переполнение невозможно)
	if len(token) >= 8 {
		digits := tokenToDigits(token[0:8])
		if !checkDigits(digits) {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		u64 += parseDigits(digits)
		i += 8
		if len(token) >= 16 {
			digits := tokenToDigits(token[8:16])
			if !checkDigits(digits) {
				return 0, &IntError{string(orig), strconv.ErrSyntax}
			}
			u64 = u64*1e8 + parseDigits(digits)
			i += 8
		}
	}

	// парсим хвостик (осталось не более 7 цифр)
	for ; i < len(token); i++ { // переполнение невозможно (20-я цифра спрятана)
		digit := token[i] - '0'
		if digit > 9 {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		u64 = u64*10 + uint64(digit)
	}
	if n == 20 { // проверяем на переполнение только здесь
		if twentiethDigit > 9 {
			return 0, &IntError{string(orig), strconv.ErrSyntax}
		}
		if u64 < math.MaxUint64/10 || (u64 == math.MaxUint64/10 && twentiethDigit <= math.MaxUint64%10) {
			u64 = u64*10 + uint64(twentiethDigit)
		} else {
			return 0, &IntError{string(orig), strconv.ErrRange}
		}
	}

	if unsigned {
		if u64 > uint64(^T(0)) {
			return 0, &IntError{string(orig), strconv.ErrRange}
		}
		return T(u64), nil
	}

	// signed range check
	bits := int(unsafe.Sizeof(T(0))) << 3
	absMin := uint64(1) << (bits - 1) // |min(T)|
	if orig[0] == '-' {
		if u64 > absMin {
			return 0, &IntError{string(orig), strconv.ErrRange}
		}
		return -T(u64), nil // two's complement: корректно при u64==absMin
	}
	if u64 >= absMin {
		return 0, &IntError{string(orig), strconv.ErrRange}
	}
	return T(u64), nil
}
