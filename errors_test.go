package executors

import (
	"fmt"
	"testing"
)

func TestErrors(t *testing.T) {
	if errs := getErrors(); errs != nil {
		fmt.Printf("empty dangerErrorList is not nil\n")
	} else {
		t.Errorf("empty dangerErrorList is nil")
	}
}

func getErrors() error {
	var errs errorList
	return errs
}

func TestAppendError(t *testing.T) {
	err1 := fmt.Errorf("err1")
	err2 := fmt.Errorf("err2")
	err3 := fmt.Errorf("err3")
	err4 := fmt.Errorf("err4")
	cases := []struct {
		in   []error
		want int
	}{
		{[]error{nil}, 0},
		{[]error{nil, nil}, 0},
		{[]error{nil, nil, nil}, 0},
		{[]error{nil, nil, nil, nil}, 0},
		{[]error{err1}, 1},
		{[]error{err1, nil}, 1},
		{[]error{nil, err1}, 1},
		{[]error{err1, err2}, 2},
		{[]error{err1, err2, nil}, 2},
		{[]error{nil, err1, err2}, 2},
		{[]error{err1, nil, err2}, 2},
		{[]error{err1, err2, err3}, 3},
		{[]error{err1, err2, err3, nil}, 3},
		{[]error{nil, err1, err2, err3}, 3},
		{[]error{err1, nil, err2, err3}, 3},
		{[]error{err1, err2, nil, err3}, 3},
		{[]error{err1, err2, err3, err4}, 4},
		{[]error{err1, err2, err3, err4, nil}, 4},
		{[]error{nil, nil, err2, err3, err4}, 3},
		{[]error{err1, nil, nil, err3, err4}, 3},
		{[]error{err1, err2, nil, nil, err4}, 3},
		{[]error{err1, err2, err3, nil, nil}, 3},
		{[]error{err1, err2, err3, err4, nil}, 4},
		{[]error{nil, errorList{err1, err2}, errorList{err3, err4}}, 4},
	}
	for _, c := range cases {
		got := AppendError(c.in...)
		if getErrorLen(got) != c.want {
			t.Errorf("AppendError(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func getErrorLen(err error) int {
	switch v := err.(type) {
	case nil:
		return 0
	case errorList:
		return len(v)
	case unwrapError:
		return len(v.Unwrap())
	default:
		return 1
	}
}

func TestRuntimeError(t *testing.T) {
	err1 := fmt.Errorf("fmt")
	if IsRuntimeError(err1) {
		t.Errorf("err1 should not_be runtimeError")
	}
	err2 := NewRuntimeErrorf("runtimeError")
	if !IsRuntimeError(err2) {
		t.Errorf("err2 should be runtimeError")
	}
	err3 := AppendError(err1, err2)
	if !IsRuntimeError(err3) {
		t.Errorf("err3 should be runtimeError")
	}
	err4 := AppendError(err1, err1)
	if IsRuntimeError(err4) {
		t.Errorf("err4 should not_be runtimeError")
	}
}
