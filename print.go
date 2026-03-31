package contestio

type writeOpts = WO

// WO (write options) задаёт параметры форматирования при выводе.
type WO struct {
	Begin string // строка, выводимая перед первым элементом.
	Sep   string // разделитель между элементами.
	End   string // строка, выводимая после последнего элемента.
}

type printValFunc[T any] func(bw *Writer, v T) error
type appendValFunc[T any] func([]byte, T) []byte

func writeAppendFunc[T any](bw *Writer, appendVal appendValFunc[T], v T) (err error) {
	if bw.Available() < len(bw.scratch) {
		_, err = bw.Write(appendVal(bw.scratch[:0], v))
	} else {
		_, err = bw.Write(appendVal(bw.AvailableBuffer(), v))
	}
	return
}

func printSliceCommon[T any](bw *Writer, op writeOpts, printVal printValFunc[T], a []T) (int, error) {
	var err error

	_, _ = bw.WriteString(op.Begin)

	for i := 0; i < len(a); i++ {
		if i > 0 {
			bw.WriteString(op.Sep)
		}
		if err = printVal(bw, a[i]); err != nil {
			return i, err
		}
	}

	_, err = bw.WriteString(op.End)
	return len(a), err
}

// TODO: принять решение: производительность vs объем поддерживаемого кода

// func printNums[T Int](bw *Writer, op writeOpts, appendVal appendValFunc[T], a []T) (int, error) {
// 	_, _ = bw.WriteString(op.Begin)
// 	var buf []byte
// 	for i := 0; i < len(a); i++ {
// 		if bw.Available() < len(bw.scratch) {
// 			buf = bw.scratch[:0]
// 		} else {
// 			buf = bw.AvailableBuffer()
// 		}
// 		if i > 0 {
// 			buf = append(buf, op.Sep...)
// 		}
// 		buf = appendVal(buf, a[i])
// 		if _, err := bw.Write(buf); err != nil {
// 			return i, err
// 		}
// 	}
// 	_, err := bw.WriteString(op.End)
// 	return len(a), err
// }

// func printStrings[T ~string](bw *Writer, op writeOpts, a []T) (int, error) {
// 	_, _ = bw.WriteString(op.Begin)
// 	for i := 0; i < len(a); i++ {
// 		if i > 0 {
// 			_, _ = bw.WriteString(op.Sep)
// 		}
// 		if _, err := bw.WriteString(string(a[i])); err != nil {
// 			return i, err
// 		}
// 	}
// 	_, err := bw.WriteString(op.End)
// 	return len(a), err
// }

// func printAnyValues(bw *Writer, op writeOpts, a []any) (int, error) {
// 	_, _ = bw.WriteString(op.Begin)
// 	for i := 0; i < len(a); i++ {
// 		if i > 0 {
// 			_, _ = bw.WriteString(op.Sep)
// 		}
// 		printAny(bw, a[i])
// 	}
// 	_, err := bw.WriteString(op.End)
// 	return len(a), err
// }

var lineWO = WO{Sep: " ", End: "\n"}

func printSlice[T any](bw *Writer, op writeOpts, printVal printValFunc[T], a []T) (int, error) {
	return must(printSliceCommon(bw, op, printVal, a))
}

func printSliceLn[T any](bw *Writer, printVal printValFunc[T], a []T) (int, error) {
	return must(printSliceCommon(bw, lineWO, printVal, a))
}

func printVals[T any](bw *Writer, op writeOpts, printVal printValFunc[T], a ...T) (int, error) {
	return must(printSliceCommon(bw, op, printVal, a))
}

func printValsLn[T any](bw *Writer, printVal printValFunc[T], a ...T) (int, error) {
	return must(printSliceCommon(bw, lineWO, printVal, a))
}
