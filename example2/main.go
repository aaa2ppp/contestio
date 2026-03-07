package main

import (
	"fmt"
	"io"
	"log"
	"os"

	. "github.com/aaa2ppp/contestio"
)

// Пример ВВОДА/ВЫВОДА (с "сахарком")
//
// На первой строке задано целое число N <= 10^6 - размер массива A. Если N = -1, то размер
// массива A не задан. Следующая строка содержит элементы массива A, целые числа по модулю
// не превышающие 10^9 разделенные пробелами.
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
	var a Ints[int32]
	var ans Ints[int64]

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
		a = Resize(a, n)
		if i, err := ScanSlice(br, a); err != nil {
			return fmt.Errorf("scan a[%d]: %v", i, err)
		}
	} else { // Количество не указано (-1), читаем до конца строки.
		// Имеет смысл преаллоцировать по максимуму.
		// Если в задаче указам лимит, то гарантировано будут такие тесты.
		// Но можно указать 0 или не преаллоцировать вовсе, а передать в ScanIntsLn nil.
		const prealloc = 1 << 20 // 1M

		var err error // Декларируем err отдельно, если в следующей строке использовать := это затенит `a`
		a, err = ScanSliceLn(br, Grow(a, prealloc))
		if err != nil && err != io.EOF {
			return fmt.Errorf("scan a[%d]: %v", len(a), err)
		}
	}
	if debug {
		log.Printf("a: %v", a)
	}

	ans = solve(a)

	if i, err := PrintSliceLn(bw, ans); err != nil {
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

func init() {
	// Кастомизируем вывод логгера (дата время нас не интересует, а строчка в коде - важно)
	log.SetFlags(log.Llongfile)
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("run: %v", err)
	}
}
