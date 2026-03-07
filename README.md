# Contest IO — быстрый ввод/вывод для алгоритмических задач на Go

`contestio` — это набор функций и обёртки над `bufio.Reader` и `bufio.Writer`, которые позволяют читать и выводить данные с минимальными аллокациями и скоростью, близкой к ручному коду. Библиотека родилась из усталости писать одни и те же циклы для чтения массивов, разочарования в производительности `fmt` и желания иметь удобный инструмент, совместимый с ручным парсингом.

## Мотивация

В алгоритмических задачах часто нужно быстро прочитать миллионы чисел или строк. Стандартный `fmt.Fscan*` 
удобен, но:
- Медленный из-за рефлексии и огромного количества аллокаций (до нескольких миллионов на 1e6 чисел).
- Каждое число преобразуется в строку, создавая лишнюю нагрузку на GC.
- Чтение массива известной длины требует цикла с отдельным вызовом `fmt.Fscan` на каждой итерации.

`bufio.Scanner` с `ScanWords` даёт zero-allocation при разбиении на токены, но его API несовместим с `fmt.Scan` — нельзя читать и числа, и строки вперемешку без костылей.

Поэтому я написал свой токенизатор поверх `bufio.Reader`, который:
- Работает с **нулевыми аллокациями** — внутренние операции (пропуск пробелов, выделение токена) не создают новых объектов. Память выделяется только под буферы `bufio.Reader`/`bufio.Writer` и под слайс результатов, если вы его не преаллоцировали.
- Позволяет читать как одиночные значения (можно использовать в циклах), так и целые массивы за один вызов.
- Обобщён для всех целых типов (`int`, `int32`, `uint64` и т.д.) и чисел с плавающей точкой (`float32`, `float64`) через дженерики.
- Выводит данные так же эффективен, с поддержкой кастомных разделителей и окончаний строк.

## Особенности

- **Нулевые аллокации** — чтение токенов возвращает слайс, указывающий прямо во внутренний буфер `bufio.Reader`. Парсинг чисел не создаёт промежуточных строк. Все операции с уже выделенными слайсами (например, `ScanInts` в предварительно созданный слайс) не требуют дополнительных выделений памяти. Бенчмарки подтверждают: `scanSlice` и `scanVars_loop` делают 0 аллокаций на 1e6 чисел.
- **Обобщённые типы** — единые функции для всего семейства целых и единые функции для семейсва чисел с плавающией точкой.
- **Гибкость** — можно читать как заранее известное количество чисел, так и все числа из строки.
- **Скорость** — на уровне ручного кода с `strconv.ParseInt` и буферизацией. В 5–10 раз быстрее `fmt.Fscan` на больших объёмах.
- **Совместимость** — можно использовать точечное чтение внутри циклов, комбинировать с ручным вызовом `br.ReadString('\n')` для строк, fmt.Fscan и т.д.
- **Простой вывод** — функции `PrintIntsLn`, `PrintFloatLn` и аналоги с опциями форматирования.


## Примеры использования

### Чтение одного числа
```go
br := NewReader(os.Stdin)
var n int
ScanIntLn(br, &n) // читает строку с одним числом
```

### Чтение массива известной длины
```go
var n int
ScanIntLn(br, &n)
a := make([]int32, n)
ScanInts(br, a) // заполнит слайс (0 аллокаций после создания слайса)
```

### Чтение массива неизвестной длины (до конца строки)
```go
a, _ := ScanIntsLn(br, nil) // вернёт слайс со всеми числами строки (аллокации только под слайс)
```

### Вывод массива
```go
ans := []int64{1, 4, 9}
PrintIntsLn(bw, ans) // "1 4 9\n" (0 аллокаций)
```

### Кастомное форматирование
```go
opts := cpio.WO{Begin: "[", Sep: ", ", End: "]\n"}
PrintInts(bw, opts, ans) // "[1, 4, 9]\n"
```

### Совместимость с ручным чтением
```go
// Читаем число, затем строку
var x int
ScanInt(br, &x)
line, _ := br.ReadString('\n') // остаток строки
```

### Пример решения

[main.go](./example/main.go)

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

### Типы

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

## Производительность

Все бенчмарки проведены на 1e6 чисел (int или float64). Тестировались разные подходы:

- **fmt.Fscan / fmt.Fprintf** — стандартные функции.
- **ReadString** — чтение строки и разбиение через `strings.Fields` + `strconv`.
- **WordScanner** — `bufio.Scanner` со `ScanWords` + `strconv`.
- **scanSlice / scanVars_loop** — corner функции этой библиотеки.
- **AppendInt/AppendFloat** — ручной вывод через `strconv.AppendInt` в буфер.
- **printSlice / printVals_loop** — corner функции библиотеки для вывода.

### Чтение (1e6 int)

```
Benchmark_scanInt/fmt.Fscan-4                  2    618985400 ns/op   19762584 B/op   1045953 allocs/op
Benchmark_scanInt/ReadString-4                12     89693625 ns/op   55760872 B/op      4732 allocs/op
Benchmark_scanInt/WordScanner-4                7    144254757 ns/op          0 B/op         0 allocs/op
Benchmark_scanInt/scanSlice-4                 13     79979915 ns/op          0 B/op         0 allocs/op
Benchmark_scanInt/scanSliceLn-4               15     77500127 ns/op          0 B/op         0 allocs/op
Benchmark_scanInt/scanVars_loop-4             13     83494254 ns/op          0 B/op         0 allocs/op
```

### Чтение (1e6 float64)

```
Benchmark_scanFloat/fmt.Fscan-4                2    835818900 ns/op   24797432 B/op   1048580 allocs/op
Benchmark_scanFloat/ReadString-4               7    151933486 ns/op   61077480 B/op      5363 allocs/op
Benchmark_scanFloat/WordScanner-4              5    210770760 ns/op          0 B/op         0 allocs/op
Benchmark_scanFloat/scanSlice-4                8    139248838 ns/op          0 B/op         0 allocs/op
Benchmark_scanFloat/scanSliceLn-4              7    143293514 ns/op          0 B/op         0 allocs/op
Benchmark_scanFloat/scanVars_loop-4            7    147450843 ns/op          0 B/op         0 allocs/op
```

### Вывод (1e6 int)

```
Benchmark_printInt/fmt.Fprintf-4               10    108052750 ns/op   7865136 B/op   983055 allocs/op
Benchmark_printInt/AppendInt-4                 30     39034173 ns/op    101977 B/op     3958 allocs/op
Benchmark_printInt/AppendInt_scr-4             27     39782230 ns/op        32 B/op        1 allocs/op
Benchmark_printInt/printSlice-4                26     42207773 ns/op         0 B/op        0 allocs/op
Benchmark_printInt/printVals_loop-4            21     53772014 ns/op         0 B/op        0 allocs/op
```

### Вывод (1e6 float64)

```
Benchmark_printFloat/fmt.Fprintf-4              6    198513983 ns/op   8389177 B/op  1048578 allocs/op
Benchmark_printFloat/AppendFloat-4             10    110002810 ns/op    183376 B/op     7923 allocs/op
Benchmark_printFloat/AppendFloat_scr-4          9    111117833 ns/op        32 B/op        1 allocs/op
Benchmark_printFloat/printSlice-4               9    113114544 ns/op         0 B/op        0 allocs/op
Benchmark_printFloat/printVals_loop-4           9    125123011 ns/op         0 B/op        0 allocs/op
```

**Выводы:**
- `scanSlice` и `scanVars_loop` работают в 6–8 раз быстрее `fmt.Fscan` и не создают аллокаций (если слайс заранее выделен).
- `printSlice` и `printVals_loop` дают нулевое выделение памяти и в 2–3 раза быстрее `fmt.Fprintf`.
- По сравнению с `bufio.Scanner` (`WordScanner`) библиотека выигрывает за счёт меньшего количества вызовов и отсутствия промежуточных аллокаций при парсинге (в `WordScanner` тоже 0 аллокаций, но он медленнее из-за особенностей реализации).

## Как это работает

- **Чтение:** `bufio.Reader` буферизует входной поток. Функции `nextToken` и `skipSpace` работают напрямую с буфером, не копируя данные — возвращают слайс, ссылающийся на внутренний массив.
- **Парсинг чисел:** для целых используется быстрый путь через `strconv.Atoi` (если тип помещается в `int`) или `ParseInt`/`ParseUint` для остальных. Никаких промежуточных строк.
- **Вывод:** `strconv.AppendInt`/`AppendFloat` записывают число прямо в буфер `bufio.Writer` (или во внутренний `scratch` для оптимизации), избегая создания строк и паразитных аллокаций.
