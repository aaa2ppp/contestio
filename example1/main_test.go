package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func Test_run(t *testing.T) {
	// NOTE: Сейчас здесь мы больше тестируем поведение функций ScanXXX/PrintXXX.
	// В реальных задачах корректный ввод нам гарантируют. Удобнее сразу падать
	// на любых ошибка ввода/вода, а здесь проверять только решение примеров.

	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
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
			"bad input input",
			args{strings.NewReader("1 2 3 abc 5 6")},
			"",
			true,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Verbose() {
				defer func(f bool) { debug = f }(debug)
				debug = tt.debug
			}

			w := &bytes.Buffer{}
			if err := run(tt.args.r, w); (err != nil) != tt.wantErr {
				t.Errorf("run1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotW := strings.TrimSpace(w.String())
			if gotW != tt.wantW {
				t.Errorf("run1() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
