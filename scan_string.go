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
func ScanBytes(r *Reader, delim byte) ([]byte, error) { return _must(_scanBytes(r, delim)) }

func _scanBytes(r *Reader, delim byte) ([]byte, error) {
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
func ScanString(r *Reader, delim byte) (string, error) { return _must(_scanString(r, delim)) }

func _scanString(r *Reader, delim byte) (string, error) {
	b, err := _scanBytes(r, delim)
	return _unsafeString(b), err
}
