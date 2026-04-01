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

func printSliceCommon[T any](bw *Writer, op writeOpts, printVal printValFunc[T], a []T) (int, error) {
	_, _ = bw.WriteString(op.Begin)
	for i := range a {
		if i > 0 {
			bw.WriteString(op.Sep)
		}
		if err := printVal(bw, a[i]); err != nil {
			return i, err
		}
	}
	_, err := bw.WriteString(op.End)
	return len(a), err
}

func writeAppendFunc[T any](bw *Writer, appendVal appendValFunc[T], v T) (err error) {
	if bw.Available() < len(bw.scratch) {
		_, err = bw.Write(appendVal(bw.scratch[:0], v))
	} else {
		_, err = bw.Write(appendVal(bw.AvailableBuffer(), v))
	}
	return
}

func printSliceAppendCommon[T any](bw *Writer, op writeOpts, appendVal appendValFunc[T], a []T) (int, error) {
	var buf []byte
	_, _ = bw.WriteString(op.Begin)
	for i := range a {
		if bw.Available() < len(bw.scratch) {
			buf = bw.scratch[:0]
		} else {
			buf = bw.AvailableBuffer()
		}
		if i > 0 {
			buf = append(buf, op.Sep...)
		}
		buf = appendVal(buf, a[i])
		if _, err := bw.Write(buf); err != nil {
			return i, err
		}
	}
	_, err := bw.WriteString(op.End)
	return len(a), err
}

var lineWO = WO{Sep: " ", End: "\n"}

func printSlice[T any](bw *Writer, op writeOpts, printVal printValFunc[T], a []T) (int, error) {
	return must(printSliceCommon(bw, op, printVal, a))
}

func printSliceAppend[T any](bw *Writer, op writeOpts, appenVal appendValFunc[T], a []T) (int, error) {
	return must(printSliceAppendCommon(bw, op, appenVal, a))
}
