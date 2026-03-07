package main

import (
	"fmt"
	"io"
	"log"
	"os"

	. "github.com/aaa2ppp/contestio"
)

// Пример ВВОДА/ВЫВОДА
//
// На первой строке задано целое число N <= 1e6 - размер массива A. Если N = -1, то размер
// массива A не задан. Следующая строка содержит элементы массива A, целые числа по модулю
// не превышающие 1e9 разделенные пробелами.
//
// Выведите квадраты всех элементов массива A разделенные пробелами в том же порядке в котором
// они заданы.

func solve(a []int32) []int64 {
	ans := make([]int64, len(a))
	for i := range a {
		v := int64(a[i])
		ans[i] = v * v
	}
	return ans
}

func run(in io.Reader, out io.Writer) error {
	br := NewReader(in)
	bw := NewWriter(out)
	defer bw.Flush()

	var n int
	var a []int32

	if _, err := ScanIntLn(br, &n); err != nil {
		// Ошибка возвращается только для проверки ввода/вывода.
		// В алгоритмичиских задачах ввод нам гарантируют, по этому на ошибках ввода
		// удобнее падать log.Fatal(err). Если падаем на примерах - значит не понял условие.
		return fmt.Errorf("scan n: %v", err)
	}
	if debug {
		log.Printf("n: %d\n", n)
	}

	if n >= 0 {
		a = make([]int32, n)
		if i, err := ScanInts(br, a); err != nil {
			return fmt.Errorf("scan a[%d]: %v", i, err)
		}
	} else { // количество не указано (-1), читаем до конца строки
		// Имеет смысл преаллоцировать по максимуму.
		// Если в задаче указам лимит, то гарантировано будут такие тесты.
		// Но можно указать 0 или не преаллоцировать вовсе, а передать в ScanIntsLn nil.
		const prealloc = 1 << 20       // 1M
		a = make([]int32, 0, prealloc) // важно утановить для `a` len=0, ScanIntsLn добавляет элементы к слайсу

		var err error                  // декларируем err отдельно, если в следующей строчке использовать := это затенит `a`
		a, err = ScanIntsLn(br, a[:0]) // еще раз, ВАЖНО утановить для `a` len=0, ScanIntsLn ДОБАВЛЯЕТ элементы к слайсу
		if err != nil && err != io.EOF {
			return fmt.Errorf("scan a[%d]: %v", len(a), err)
		}
	}
	if debug {
		log.Printf("a: %v", a)
	}

	ans := solve(a)

	if i, err := PrintIntsLn(bw, ans); err != nil {
		// Ошибка возвращается только для проверки ввода/вывода.
		// В алгоритмических задачах обрабатывать эту ошибку не имеет смысла.
		// 99.9% это проблемы io os.
		return fmt.Errorf("print ans[%d]: %v", i, err)
	}

	return nil
}

// Устанавливает флаг debug режима, если определена переменная окружения DEBUG.
//
// Пример использования (bash) (значение может быть любым и отсутствовать, как в примере):
//
//	$ DEBUG= go run main.go
var _, debug = os.LookupEnv("DEBUG")

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
