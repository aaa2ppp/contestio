package contestio

import (
	"errors"
)

// Abs возвращает абсолютное значение целого числа.
// Для знаковых паникует, если аргумент равен минимальному значению типа (например, -128 для int8),
// так как результат не помещается в тот же знаковый тип.
func Abs[T Int](a T) T {
	if a < 0 {
		if a == -a {
			panic("Abs: overflow for minimum signed value")
		}
		return -a
	}
	return a
}

// GCD возвращает наибольший общий делитель (НОД).
// Возвращает 0, если оба аргумента равны 0.
func GCD[T Int](a, b T) T {
	a, b = Abs(a), Abs(b)
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// GCDx – расширенный алгоритм Евклида.
// Возвращает НОД и коэффициенты x, y: a*x + b*y = gcd(a,b).
// Возвращает (0, 1, 0), если оба аргумента равны 0.
func GCDx[T Signed](a, b T) (d, x, y T) {
	a, b = Abs(a), Abs(b)
	x0, x1 := T(1), T(0)
	y0, y1 := T(0), T(1)
	for b != 0 {
		q := a / b
		a, b = b, a-q*b
		x0, x1 = x1, x0-q*x1
		y0, y1 = y1, y0-q*y1
	}
	return a, x0, y0
}

// ModInv возвращает значение обратное `a` по модулю, или ошибку, если `a` и `mod` не являются взаимно простыми.
func ModInv[T Signed](a, mod T) (T, error) {
	d, x, _ := GCDx(a, mod)
	if d != 1 {
		return 0, errors.New("arguments are not mutually prime")
	}
	if x < 0 {
		x += mod
	}
	return T(x), nil
}
