package executors

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestParallel(t *testing.T) {
	fn := Parallel(func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		i, err := Cast[int](s)
		if err != nil {
			return s, err
		}
		return i + 100, nil
	}, func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		i, err := Cast[int](s)
		if err != nil {
			return s, err
		}
		return i + 200, nil
	}, func(s Stateful) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return s, errors.New("random_error")
		}
		i, err := Cast[int](s)
		if err != nil {
			return s, err
		}
		return i + 300, nil
	})
	sin := 1
	outs, errCount, err := RunTest(fn, sin)
	if err != nil {
		t.Error(err)
	}
	ov, ok := outs.([]any)
	if !ok {
		t.Errorf("outs: %v, ty: %v", outs, reflect.TypeOf(outs))
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(ov), errCount)
}

func TestParallelStruct(t *testing.T) {
	type Counter struct {
		Count int
	}
	fn := Parallel(func(sin Counter) (Stateful, error) {
		sin.Count++
		if sin.Count < 10 {
			return sin, fmt.Errorf("error: %d", sin.Count)
		}
		return "succeed", nil
	})
	var sin any
	outs, errCount, err := RunTest(fn, sin)
	if err != nil {
		t.Error(err)
	}
	ov, ok := outs.([]any)
	if !ok {
		t.Errorf("outs: %v, ty: %v", outs, reflect.TypeOf(outs))
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(ov), errCount)
}
