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
	err := getErrorN(0)
	if err != nil {
		t.Errorf("err0 is not nil: %v", err)
	} else {
		fmt.Printf("err0: %v\n", err)
	}
	for i := 1; i < 5; i++ {
		err := getErrorN(i)
		if err == nil {
			t.Errorf("err%d is nil: %v", i, err)
		} else {
			fmt.Printf("err%d: %v\n", i, err)
		}
	}
}

func getErrorN(n int) error {
	var errOut error
	for i := 0; i < n; i++ {
		errOut = AppendError(errOut, fmt.Errorf("err%d", i))
	}
	return errOut
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
