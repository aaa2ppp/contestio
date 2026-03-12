//go:build sugar

package contestio

import (
	"fmt"
	"reflect"
	"testing"
)

func TestResize(t *testing.T) {
	type args struct {
		s []int
		n int
	}
	tests := []struct {
		name string // description of this test case
		args
		want      []int
		wantX     []int
		wantPanic bool
	}{
		{
			"grow",
			args{[]int{1, 2, 3, 4, 5}, 8},
			[]int{1, 2, 3, 4, 5, 0, 0, 0},
			nil,
			false,
		},
		{
			"grow empty",
			args{[]int{}, 5},
			[]int{0, 0, 0, 0, 0},
			nil,
			false,
		},
		{
			"grow nil",
			args{nil, 5},
			[]int{0, 0, 0, 0, 0},
			nil,
			false,
		},
		{
			"trim",
			args{[]int{1, 2, 3, 4, 5}, 3},
			[]int{1, 2, 3},
			[]int{1, 2, 3, 0, 0},
			false,
		},
		{
			"trim empty",
			args{[]int{}, 0},
			[]int{},
			nil,
			false,
		},
		{
			"trim nil",
			args{nil, 0},
			nil,
			nil,
			false,
		},
		{
			"garbage after len",
			args{[]int{1, 2, 3, 4, 5}[:3], 1},
			[]int{1},
			[]int{1, 0, 0, 4, 5},
			false,
		},
		{
			"negative size",
			args{[]int{1, 2, 3, 4, 5}, -1},
			nil,
			nil,
			true,
		},
		{
			"no change",
			args{[]int{1, 2, 3}, 3},
			[]int{1, 2, 3},
			[]int{1, 2, 3},
			false,
		},
	}

	catchPanic := func(fn func([]int, int) []int, s []int, n int) (_ []int, err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("*** panic ***: %v", p)
			}
		}()
		return fn(s, n), nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := catchPanic(Resize, tt.s, tt.n)
			if (err != nil) != tt.wantPanic {
				t.Fatalf("unexpected: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if tt.wantX != nil {
				if gotX := got[0:cap(got)]; !reflect.DeepEqual(gotX, tt.wantX) {
					t.Errorf("gotX %v, wantX %v", gotX, tt.wantX)
				}
			}
		})
	}
}
