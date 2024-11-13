package executors

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
)

func mockStatefulTask() (TaskFunc, Stateful) {
	fn := Pipe(func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		return fmt.Sprintf("mock: %v", s), nil
	}, func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		return fmt.Sprintf("mock: %v", s), nil
	}, func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		return fmt.Sprintf("mock: %v", s), nil
	})
	sin := fmt.Sprintf("mock%d", 1)
	return fn, sin
}

func TestPipe(t *testing.T) {
	fn1, s1 := mockStatefulTask()
	fn := Pipe(func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		return fmt.Sprintf("pipe1: %v", s), nil
	}, func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		return fmt.Sprintf("pipe2: %v", s), nil
	}, func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		return s1, nil
	}, fn1)
	outs, errCount, err := RunTest(fn, 1)
	if err != nil {
		t.Error(err)
	}
	os, ok := outs.(string)
	if !ok {
		t.Errorf("outs: %v, ty: %v", outs, reflect.TypeOf(outs))
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(os), errCount)
}
