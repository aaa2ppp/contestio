package contestio

import (
	"bytes"
	"io"
)

// ScanBytes читает до первого вхождения delim включительно или конца файла (EOF).
// При успешном чтении возвращает срез байт без delim, из которого удалены
// завершающие пробельные символы (пробел, табуляция, \r, \n).
// io.EOF возвращается только в том случае, если не было прочитано ни одного байта.
// При любой другой ошибке возвращаются все прочитанные данные и эта ошибка.
func ScanBytes(r *Reader, delim byte) ([]byte, error) { return must(scanBytes(r, delim)) }

func scanBytes(r *Reader, delim byte) ([]byte, error) {
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

// ScanString читает до первого вхождения delim включительно или конца файла (EOF).
// При успешном чтении возвращает строку без delim, из которой удалены
// завершающие пробельные символы (пробел, табуляция, \r, \n).
// io.EOF возвращается только в том случае, если не было прочитано ни одного байта.
// При любой другой ошибке возвращаются все прочитанные данные и эта ошибка.
func ScanString(r *Reader, delim byte) (string, error) { return must(scanString(r, delim)) }

func scanString(r *Reader, delim byte) (string, error) {
	b, err := scanBytes(r, delim)
	return unsafeString(b), err
}
