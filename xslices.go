package contestio

// Пара очень удобных фунций для управления размером (len) и емкостью (cap) слайсов.
// Обратите внимание эти функции используют семантику append.

// Grow возвращает слайс в который можно добавить не менее n элементов без
// дополнительных аллокаций.
func Grow[T any, S ~[]T](s S, n int) S {
	if cap(s)-len(s) >= n {
		return s
	}
	newCap := cap(s) * 2
	if newCap < len(s)+n {
		newCap = len(s) + n
	}
	newSlice := make([]T, len(s), newCap)
	copy(newSlice, s)
	return newSlice
}

// Resize возвращает слайс размера n. Если емкость исходноко слайса меньше n,
// слай будет расширен с помощью Grow, если меньше, то усечен до n с занулением
// элементов с индексом от n до конца слайса.
func Resize[T any, S ~[]T](s S, n int) S {
	if n > cap(s) {
		s = Grow(s, n-len(s))
	} else {
		clear(s[n:])
	}
	return s[:n]
}
