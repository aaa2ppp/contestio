package contestio

import (
	"bufio"
	"io"
)

const defaultBufSize = 4096

type br = bufio.Reader
type bw = bufio.Writer

// Reader буферизированный читатель.
type Reader struct{ br }

// NewReaderSize создаёт новый Reader с заданным размером буфера.
func NewReaderSize(r io.Reader, size int) *Reader {
	return &Reader{*bufio.NewReaderSize(r, size)}
}

// NewReader создаёт новый Reader с размером буфера по умолчанию (4096).
func NewReader(r io.Reader) *Reader {
	return NewReaderSize(r, defaultBufSize)
}

// Writer буферизированный писатель.
type Writer struct {
	bw
	scratch [32]byte
}

// NewWriterSize создаёт новый Writer с заданным размером буфера.
func NewWriterSize(w io.Writer, size int) *Writer {
	return &Writer{bw: *bufio.NewWriterSize(w, size)}
}

// NewWriter создаёт новый Writer с размером буфера по умолчанию (4096).
func NewWriter(w io.Writer) *Writer {
	return NewWriterSize(w, defaultBufSize)
}
