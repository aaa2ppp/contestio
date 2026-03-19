package contestio

import (
	"bufio"
	"bytes"
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

func (r *Reader) readBytes(delim byte) ([]byte, error) {
	b, err := r.br.ReadBytes(delim)
	if err != nil && err != io.EOF {
		return b, err
	}
	if len(b) == 0 {
		return nil, io.EOF
	}
	if err == nil {
		b = b[:len(b)-1]
	}
	return bytes.TrimRight(b, " \t\r\n"), nil
}

func (r *Reader) readString(delim byte) (string, error) {
	b, err := r.readBytes(delim)
	return unsafeString(b), err
}

// ReadBytes reads until the first occurrence of delim in the input,
// returning a slice of the read data with the delimiter removed
// and trailing whitespace (space, tab, \r, \n) trimmed.
// The returned slice may point into the internal buffer;
// the trimmed part is still available in the slice's capacity.
// If delim is encountered, the returned slice does not include delim.
// If delim is not encountered before EOF, all remaining data is returned
// (after trimming trailing whitespace) and the error is nil.
// If no data was read at all (empty input), it returns nil, io.EOF.
// If a non-EOF error occurs, it returns the data read so far and the error.
func (r *Reader) ReadBytes(delim byte) ([]byte, error) { return r.readBytes(delim) }

// ReadString reads until the first occurrence of delim in the input,
// returning a string of the read data with the delimiter removed
// and trailing whitespace (space, tab, \r, \n) trimmed.
// The semantics are otherwise identical to ReadBytes.
func (r *Reader) ReadString(delim byte) (string, error) { return r.readString(delim) }

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
