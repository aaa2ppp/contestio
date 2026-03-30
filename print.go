package contestio

type writeOpts = WO

// WO (write options) задаёт параметры форматирования при выводе.
type WO struct {
	Begin string // строка, выводимая перед первым элементом.
	Sep   string // разделитель между элементами.
	End   string // строка, выводимая после последнего элемента.
}

type printFunc[T any] func(bw *Writer, v T) error
type appendValFunc[T any] func([]byte, T) []byte

func printVal[T any](bw *Writer, appendVal appendValFunc[T], v T) error {
	var buf []byte
	if bw.Available() < len(bw.scratch) {
		buf = bw.scratch[:0]
	} else {
		buf = bw.AvailableBuffer()
	}
	buf = appendVal(buf, v)
	if _, err := bw.Write(buf); err != nil {
		return err
	}
	return nil
}

func printSliceCommon[T any](bw *Writer, op writeOpts, print printFunc[T], a []T) (int, error) {
	var err error

	_, _ = bw.WriteString(op.Begin)

	for i := 0; i < len(a); i++ {
		if i > 0 {
			bw.WriteString(op.Sep)
		}
		if err = print(bw, a[i]); err != nil {
			return i, err
		}
	}

	_, err = bw.WriteString(op.End)
	return len(a), err
}

var lineWO = WO{Sep: " ", End: "\n"}

func printSlice[T any](bw *Writer, op writeOpts, printVal printFunc[T], a []T) (int, error) {
	return must(printSliceCommon(bw, op, printVal, a))
}

func printSliceLn[T any](bw *Writer, printVal printFunc[T], a []T) (int, error) {
	return must(printSliceCommon(bw, lineWO, printVal, a))
}

func printVals[T any](bw *Writer, op writeOpts, printVal printFunc[T], a ...T) (int, error) {
	return must(printSliceCommon(bw, op, printVal, a))
}

func printValsLn[T any](bw *Writer, printVal printFunc[T], a ...T) (int, error) {
	return must(printSliceCommon(bw, lineWO, printVal, a))
}
