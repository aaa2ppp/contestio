package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func Test_run(t *testing.T) {
	// NOTE: Сейчас здесь мы больше тестируем поведение функций ScanXXX/PrintXXX.
	// В реальных задачах корректный ввод нам гарантируют. Удобнее сразу падать
	// на любых ошибка ввода/вода, а здесь проверять только решение примеров.

	type args struct {
		in io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
		debug   bool
	}{
		{
			"line of elements",
			args{strings.NewReader("-1\n1 2 3\n")},
			"1 4 9",
			false,
			true,
		},
		{
			"line of elements without eol",
			args{strings.NewReader("-1\n1 2 3")},
			"1 4 9",
			false,
			true,
		},
		{
			"empty line",
			args{strings.NewReader("-1\n")},
			"",
			false,
			true,
		},
		{
			"count elements",
			args{strings.NewReader("3\n1 2 3")},
			"1 4 9",
			false,
			true,
		},
		{
			"missing elements",
			args{strings.NewReader("3\n1 2")},
			"",
			true,
			true,
		},
		{
			"superfluous elements",
			args{strings.NewReader("3\n1 2 3 4")},
			"1 4 9",
			false,
			true,
		},
		{
			"zero elements",
			args{strings.NewReader("0\n")},
			"",
			false,
			true,
		},
		{
			"empty input",
			args{strings.NewReader("")},
			"",
			true,
			true,
		},
		{
			"bad input",
			args{strings.NewReader("6\n1 2 3 abc 5 6")},
			"",
			true,
			true,
		},
	}

	catchPanic := func(fn func(io.Reader, io.Writer) error, r io.Reader, w io.Writer) (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("*** panic ***: %v", p)
			}
		}()
		return fn(r, w)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Verbose() {
				defer func(f bool) { debug = f }(debug)
				debug = tt.debug
			}

			out := &bytes.Buffer{}
			err := catchPanic(run, tt.args.in, out)

			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotOut := strings.TrimRight(out.String(), " \t\r\n")
			if gotOut != tt.wantOut {
				t.Errorf("run() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
