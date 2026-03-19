package contestio

// WO (write options) задаёт параметры форматирования при выводе.
type WO struct {
	Begin string // строка, выводимая перед первым элементом.
	Sep   string // разделитель между элементами.
	End   string // строка, выводимая после последнего элемента.
}
type writeOpts = WO

type appendValFunc[T any] func([]byte, T) []byte

func printSliceCommon[T any](bw *Writer, op writeOpts, appendVal appendValFunc[T], a []T) (int, error) {
	var err error
	var buf []byte

	_, _ = bw.WriteString(op.Begin)

	for i := 0; i < len(a); i++ {
		if bw.Available() < len(bw.scratch) {
			buf = bw.scratch[:0]
		} else {
			buf = bw.AvailableBuffer()
		}
		if i > 0 {
			buf = append(buf, op.Sep...)
		}
		buf = appendVal(buf, a[i])
		if _, err = bw.Write(buf); err != nil {
			return i, err
		}
	}

	_, err = bw.WriteString(op.End)
	return len(a), err
}

var lineWO = WO{Sep: " ", End: "\n"}

func printSlice[T any](bw *Writer, op writeOpts, appendVal appendValFunc[T], a []T) (int, error) {
	return printSliceCommon(bw, op, appendVal, a)
}

func printSliceLn[T any](bw *Writer, appendVal appendValFunc[T], a []T) (int, error) {
	return printSliceCommon(bw, lineWO, appendVal, a)
}

func printVals[T any](bw *Writer, op writeOpts, appendVal appendValFunc[T], a ...T) (int, error) {
	return printSliceCommon(bw, op, appendVal, a)
}

func printValsLn[T any](bw *Writer, appendVal appendValFunc[T], a ...T) (int, error) {
	return printSliceCommon(bw, lineWO, appendVal, a)
}
