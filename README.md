# Contest IO — быстрый ввод/вывод для алгоритмических задач на Go

`contestio` — это набор функций и обёртки над `bufio.Reader` и `bufio.Writer`, которые позволяют читать и выводить данные с минимальными аллокациями и скоростью, близкой к ручному коду. Библиотека родилась из усталости писать одни и те же циклы для чтения массивов, разочарования в производительности `fmt` и желания иметь удобный инструмент, совместимый с ручным парсингом.

## Мотивация

В алгоритмических задачах часто нужно быстро прочитать миллионы чисел или строк. Стандартный `fmt.Fscan*` 
удобен, но:
- Медленный из-за рефлексии и огромного количества аллокаций.
- Каждое число преобразуется в строку, создавая лишнюю нагрузку на GC.
- Чтение массива известной длины требует цикла с отдельным вызовом `fmt.Fscan` на каждой итерации.

`bufio.Scanner` с `ScanWords` даёт zero-allocation при разбиении на токены, но его API несовместим с `fmt.Scan` — нельзя читать и числа, и строки вперемешку без костылей.

Поэтому я написал свой токенизатор поверх `bufio.Reader`, который:
- Работает с **нулевыми аллокациями** — внутренние операции (пропуск пробелов, выделение токена) не создают новых объектов. Память выделяется только под буферы
`bufio.Reader`/`bufio.Writer` и под слайс результатов, если вы его не преаллоцировали.
- Позволяет читать как одиночные значения (можно использовать в циклах), так и целые массивы за один вызов.
- Обобщён для всех целых типов (`int`, `int32`, `uint64` и т.д.) и чисел с плавающей точкой (`float32`, `float64`) через дженерики.
- Выводит данные так же эффективно, с поддержкой кастомных разделителей и окончаний строк.

## Особенности

<!-- это, по моему, повторение -->
- **Нулевые аллокации** — чтение токенов возвращает слайс, указывающий прямо во внутренний буфер `bufio.Reader`. Парсинг чисел не создаёт промежуточных строк. Все операции с уже выделенными слайсами (например, `ScanInts` в предварительно созданный слайс) не требуют дополнительных выделений памяти. Бенчмарки подтверждают: `scanSlice` и `scanVars_loop` делают 0 аллокаций на 1e6 чисел.
- **Обобщённые типы** — единые функции для всего семейства целых и единые функции для семейства чисел с плавающей точкой.
- **Гибкость** — можно читать как заранее известное количество чисел, так и все числа из строки.
- **Скорость** — на уровне ручного кода с `strconv.ParseInt` и буферизацией. В 5–10 раз быстрее `fmt.Fscan` на больших объёмах.
- **Совместимость** — можно использовать точечное чтение внутри циклов, комбинировать с ручным вызовом `br.ReadString('\n')` для строк, `fmt.Fscan` и т.д.
- **Простой вывод** — функции `PrintIntsLn`, `PrintFloatLn` и аналоги с опциями форматирования.
- **Расширяемость** — легко добавить поддержку своих типов (например, строк, дат) через интерфейсы `Slice`, `Parser`, `ValAppender` (см. `word.go`).

## Установка

```bash
go get github.com/aaa2ppp/contestio/lib
```

## Примеры использования

### Простой вариант (без «сахара»)

```go
package main

import (
    . "github.com/aaa2ppp/contestio/lib"
    "os"
)

func main() {
    br := NewReader(os.Stdin)
    bw := NewWriter(os.Stdout)
    defer bw.Flush()

    var n int
    ScanIntLn(br, &n)                // читаем количество элементов

    a := make([]int, n)
    ScanInts(br, a)                   // заполняем слайс (0 аллокаций)

    // обработка ...

    PrintIntsLn(bw, a)                 // выводим результат
}
```

### Вариант с «сахаром» (типы Ints, функции Resize/Grow)

```go
package main

import (
    . "github.com/aaa2ppp/contestio/lib"
    "os"
)

func main() {
    br := NewReader(os.Stdin)
    bw := NewWriter(os.Stdout)
    defer bw.Flush()

    var n int
    ScanIntLn(br, &n)

    var a Ints[int]                    // обёртка над слайсом

    if n >= 0 {
        a = Resize(a, n)                // выделяем память
        ScanSlice(br, a)                 // читаем через интерфейс
    } else {
        // читаем все числа из строки, заранее увеличив ёмкость
        a, _ = ScanSliceLn(br, Grow(a, 1<<20))
    }

    // обработка ...

    PrintSliceLn(bw, a)                 // выводим через интерфейс
}
```

### Чтение массива неизвестной длины (до конца строки)

```go
a, _ := ScanIntsLn(br, nil)   // вернёт слайс со всеми числами строки
```

### Вывод с кастомным форматированием

```go
opts := WO{Begin: "[", Sep: ", ", End: "]\n"}
PrintInts(bw, opts, ans)      // "[1, 4, 9]\n"
```

### Совместимость с ручным чтением

```go
var x int
ScanInt(br, &x)
line, _ := br.ReadString('\n') // остаток строки (например, текст)
```

## API

### Для целых чисел (int, int8, int16, int32, int64, uint, uint8, ...)

```go
func ScanInts[T Int](br *Reader, a []T) (int, error)
func ScanIntsLn[T Int](br *Reader, a []T) ([]T, error)
func ScanInt[T Int](br *Reader, a ...*T) (int, error)
func ScanIntLn[T Int](br *Reader, a ...*T) (int, error)

func PrintInts[T Int](bw *Writer, op WO, a []T) (int, error)
func PrintIntsLn[T Int](bw *Writer, a []T) (int, error)
func PrintInt[T Int](bw *Writer, op WO, a ...T) (int, error)
func PrintIntLn[T Int](bw *Writer, a ...T) (int, error)
```

### Для чисел с плавающей точкой (float32, float64)

```go
func ScanFloats[T Float](br *Reader, a []T) (int, error)
func ScanFloatsLn[T Float](br *Reader, a []T) ([]T, error)
func ScanFloat[T Float](br *Reader, a ...*T) (int, error)
func ScanFloatLn[T Float](br *Reader, a ...*T) (int, error)

func PrintFloats[T Float](bw *Writer, op WO, a []T) (int, error)
func PrintFloatsLn[T Float](bw *Writer, a []T) (int, error)
func PrintFloat[T Float](bw *Writer, op WO, a ...T) (int, error)
func PrintFloatLn[T Float](bw *Writer, a ...T) (int, error)
```

### Для слов (любые строки не содержащие пробелов)
```go
func ScanWords[T Word](br *Reader, a []T) (int, error)
func ScanWordsLn[T Word](br *Reader, a []T) ([]T, error)
func ScanWord[T Word](br *Reader, a ...*T) (int, error)
func ScanWordLn[T Word](br *Reader, a ...*T) (int, error)

func PrintWords[T Word](bw *Writer, op WO, a []T) (int, error)
func PrintWordsLn[T Word](bw *Writer, a []T) (int, error)
func PrintWord[T Word](bw *Writer, op WO, a ...T) (int, error)
func PrintWordLn[T Word](bw *Writer, a ...T) (int, error)
```

### Для пользовательских типов

В файле [`custom_type.go`](lib/custom_type.go) показан пример создания типа `Type` (string) и соответствующих функций.
Вы можете легко адаптировать этот шаблон под свои нужды.

### Типы и конструкторы

```go
type Reader = bufio.Reader

type Writer struct {
    *bufio.Writer
    // unexported fields
}

func NewReader(r io.Reader) *Reader
func NewReaderSize(r io.Reader, size int) *Reader
func NewWriter(w io.Writer) *Writer
func NewWriterSize(w io.Writer, size int) *Writer

// WO — опции форматирования при выводе
type WO struct {
    Begin string // выводится перед первым элементом
    Sep   string // разделитель между элементами
    End   string // выводится после последнего элемента
}
```

### Вспомогательные функции для работы со слайсами

```go
// Grow увеличивает ёмкость слайса так, чтобы можно было добавить не менее n элементов.
func Grow[T any, S ~[]T](s S, n int) S

// Resize изменяет длину слайса до n, при необходимости увеличивая ёмкость.
func Resize[T any, S ~[]T](s S, n int) S
```

## Инструмент `contestio-inline` для инлайна кода

При отправке решения на проверку в системах типа Codeforces, Яндекс.Contest и т.п. нельзя подключать внешние пакеты. Чтобы обойти это ограничение, в репозитории есть утилита `cmd/contestio-inline`, которая встраивает (инлайнит) код библиотеки прямо в ваш `main.go`.

## Утилита contestio-inline

Для встраивания кода библиотеки в ваш `main.go` (например, перед отправкой на проверку) используйте утилиту `contestio-inline`.

### Установка

```bash
go install github.com/aaa2ppp/contestio/cmd/contestio-inline@latest
```

### Использование

```bash
# Перейти в директорию с решением
cd your_solution

# Инлайн кода библиотеки в main.go
contestio-inline main.go

# После этого в main.go появятся все необходимые определения из contestio,
# заключённые в специальные комментарии. Файл готов к отправке.

# Если нужно вернуть исходный вид (удалить встроенный код и восстановить импорт):
contestio-inline -clear main.go
```

Утилита анализирует, какие именно объекты библиотеки используются в вашем коде, и встраивает только достижимые определения, включая все зависимости. Это гарантирует минимальный размер итогового файла.

## Производительность

Все бенчмарки проведены на 1e6 чисел (int или float64) с буфером 4 КБ.  
**Система:** Intel Core i3-7100 @ 3.90GHz, Windows, go1.24.2.

### Чтение целых чисел (1e6 int)

| Метод              | Время (ns/op) | Память (B/op) | Аллокации |
|--------------------|---------------|---------------|-----------|
| `fmt.Fscan`        | 560 397 950   | 19 762 552    | 1 045 951 |
| `ReadString`       |  88 997 046   | 55 773 160    |     4 733 |
| `WordScanner`      | 136 198 125   |            0  |         0 |
| **`scanSlice`**    | **63 177 911**|            0  |         0 |
| **`scanSliceLn`**  | **64 200 878**|            0  |         0 |
| **`scanVars_loop`**| **67 483 444**|            0  |         0 |

### Чтение чисел с плавающей точкой (1e6 float64)

| Метод              | Время (ns/op) | Память (B/op) | Аллокации |
|--------------------|---------------|---------------|-----------|
| `fmt.Fscan`        | 755 546 700   | 24 797 432    | 1 048 580 |
| `ReadString`       | 146 385 571   | 61 077 480    |     5 363 |
| `WordScanner`      | 197 546 067   |            0  |         0 |
| **`scanSlice`**    | **134 886 212**|            0  |         0 |
| **`scanSliceLn`**  | **135 784 738**|            0  |         0 |
| **`scanVars_loop`**| **135 497 500**|            0  |         0 |

### Вывод целых чисел (1e6 int)

| Метод              | Время (ns/op) | Память (B/op) | Аллокации |
|--------------------|---------------|---------------|-----------|
| `fmt.Fprintf`      | 114 573 622   | 7 865 109     |   983 054 |
| `AppendInt`        |  39 200 129   |   101 976     |     3 958 |
| `AppendInt_scr`    |  39 807 919   |        32     |         1 |
| **`printSlice`**   | **42 321 088**|         0     |         0 |
| **`printVals_loop`**| **53 914 705**|         0     |         0 |

### Вывод чисел с плавающей точкой (1e6 float64)

| Метод              | Время (ns/op) | Память (B/op) | Аллокации |
|--------------------|---------------|---------------|-----------|
| `fmt.Fprintf`      | 201 942 660   | 8 389 315     | 1 048 579 |
| `AppendFloat`      | 113 139 067   |   183 377     |     7 923 |
| `AppendFloat_scr`  | 112 149 100   |        32     |         1 |
| **`printSlice`**   | **115 920 822**|         0     |         0 |
| **`printVals_loop`**| **126 021 100**|         0     |         0 |

**Выводы:**
- Функции `scanSlice`/`scanVars_loop` работают в **6–8 раз быстрее** `fmt.Fscan` и не создают аллокаций (если слайс заранее выделен).
- `printSlice`/`printVals_loop` дают нулевое выделение памяти и в **2–3 раза быстрее** `fmt.Fprintf`.
- По сравнению с `bufio.Scanner` (`WordScanner`) библиотека выигрывает за счёт меньшего количества вызовов и отсутствия промежуточных аллокаций при парсинге (в `WordScanner` тоже 0 аллокаций, но он медленнее из-за особенностей реализации).

## Как это работает

- **Чтение:** `bufio.Reader` буферизует входной поток. Функции `nextToken` и `skipSpace` работают напрямую с буфером, не копируя данные — возвращают слайс, ссылающийся на внутренний массив.
- **Парсинг чисел:** для целых используется быстрый путь через `strconv.Atoi` (если тип помещается в `int`) или `ParseInt`/`ParseUint` для остальных. Никаких промежуточных строк.
- **Вывод:** `strconv.AppendInt`/`AppendFloat` записывают число прямо в буфер `bufio.Writer` (или во внутренний `scratch` для оптимизации), избегая создания строк и паразитных аллокаций.

Библиотека предлагает три варианта парсера целых чисел, выбираемых тегами сборки:
- `parse_int_std` — стандартный `strconv.ParseInt`/`ParseUint`.
- `parse_int_fast` — оптимизированный парсер с ранним выходом при переполнении (используется по умолчанию).
- иначе — базовый цикл с проверками (запасной вариант).

Вы можете явно указать тег при сборке, например:
```bash
go build -tags=parse_int_std
```
