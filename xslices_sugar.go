//go:build sugar

package contestio

import "slices"

// Пара очень удобных фунций для управления размером (len) и емкостью (cap) слайсов.
// Обратите внимание эти функции используют семантику append.

// Grow возвращает слайс в который можно добавить не менее n элементов без
// дополнительных аллокаций. Паникует, если n<0.
func Grow[S ~[]T, T any](s S, n int) S { return slices.Grow(s, n) }

// Resize возвращает слайс размера n. Если емкость исходноко слайса меньше n,
// слайс будет расширен с помощью Grow, если меньше, то усечен до n с обнулением
// элементов с индексом от n до конца слайса. Паникует, если n<0.
func Resize[S ~[]T, T any](s S, n int) S {
	if n < 0 {
		panic("cannot be negative")
	}
	if n > len(s) {
		return Grow(s, n-len(s))[:n]
	}
	clear(s[n:])
	return s[:n]
}
