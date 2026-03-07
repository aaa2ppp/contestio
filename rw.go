package contestio

import (
	"bufio"
	"io"
	"unsafe"
)

const defaultBufSize = 4096

// Reader буферизированный читатель.
type Reader = bufio.Reader

// NewReaderSize создаёт новый Reader с заданным размером буфера.
func NewReaderSize(r io.Reader, size int) *Reader {
	return bufio.NewReaderSize(r, size)
}

// NewReader создаёт новый Reader с размером буфера по умолчанию (4096).
func NewReader(r io.Reader) *Reader {
	return NewReaderSize(r, defaultBufSize)
}

// Writer буферизированный писатель.
type Writer struct {
	*bufio.Writer
	scratch [64 - unsafe.Sizeof(uintptr(0))]byte // aligning the structure by class 64 bytes 8(4) + 56(60)
}

// NewWriterSize создаёт новый Writer с заданным размером буфера.
func NewWriterSize(w io.Writer, size int) *Writer {
	return &Writer{Writer: bufio.NewWriterSize(w, size)}
}

// NewWriter создаёт новый Writer с размером буфера по умолчанию (4096).
func NewWriter(w io.Writer) *Writer {
	return NewWriterSize(w, defaultBufSize)
}

