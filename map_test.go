package executors

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestMapSlice(t *testing.T) {
	fn := Map(func(sin int) (int, error) {
		if rand.IntN(100) < 50 {
			return sin, errors.New("random_error")
		}
		return sin + 100, nil
	})
	var ins Stateful = []int{0, 1, 2, 3, 4, 5, 6}
	outs, errCount, err := RunTest(fn, ins)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	ov, ok := outs.([]int)
	if !ok {
		t.Errorf("outs: %v, ty: %v", outs, reflect.TypeOf(outs))
		return
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(ov), errCount)
}

func TestMapMap(t *testing.T) {
	fn := Map(func(sin int) (int, error) {
		if rand.IntN(100) < 50 {
			return sin, errors.New("random_error")
		}
		return sin + 100, nil
	})
	var ins Stateful = map[string]int{
		"a": 0,
		"b": 1,
		"c": 2,
		"d": 3,
		"e": 4,
		"f": 5,
		"g": 6,
	}
	outs, errCount, err := RunTest(fn, ins)
	if err != nil {
		t.Error(err)
		return
	}
	ov, ok := outs.(map[string]int)
	if !ok {
		t.Errorf("outs: %v, ty: %v", outs, reflect.TypeOf(outs))
		return
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(ov), errCount)
}

func TestMapStruct(t *testing.T) {
	fn := Map(func(sin Person) (Stateful, error) {
		if rand.IntN(100) < 50 {
			return sin, errors.New("random_error")
		}
		sin.Age += 100
		return sin.Age, nil
	})
	var ins Stateful = map[string]Person{
		"a": {
			Name: "a",
			Age:  1,
		},
		"b": {
			Name: "b",
			Age:  2,
		},
		"c": {
			Name: "c",
			Age:  3,
		},
	}
	outs, errCount, err := RunTest(fn, ins)
	if err != nil {
		t.Error(err)
		return
	}
	ov, ok := outs.(map[string]any)
	if !ok {
		t.Errorf("outs: %v, ty: %v", outs, reflect.TypeOf(outs))
		return
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(ov), errCount)
}

func TestMapState(t *testing.T) {
	fn1, s1 := mockStatefulTask()
	fn := Map(fn1)
	var ins Stateful = Repeat(s1, 4)
	outs, errCount, err := RunTest(fn, ins)
	if err != nil {
		t.Error(err)
		return
	}
	ov, ok := outs.([]any)
	if !ok {
		t.Errorf("outs: %v, ty: %v", outs, reflect.TypeOf(outs))
		return
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(ov), errCount)
}

func TestMapNil(t *testing.T) {
	fn := Map(func(sin any) (any, error) {
		fmt.Printf("sin: %v", sin)
		return sin, nil
	})
	outs, errCount, err := RunTest(fn, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if outs != nil {
		t.Errorf("outs: %v, ty: %v", outs, reflect.TypeOf(outs))
		return
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(outs), errCount)
}

func TestMapMapError(t *testing.T) {
	var i = 0
	fn := Map(func(sin string) (any, error) {
		i++
		if i < 8 {
			return sin, errors.New("first_error")
		}
		return Person{
			Name: sin,
		}, nil
	})
	sin := map[string]string{
		"lb":    "lb",
		"lb_sg": "lb_sg",
	}
	out, errCount, err := RunTest(fn, sin)
	if err != nil {
		t.Error(err)
		return
	}
	outs, err := Cast[map[string]Person](out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(outs) != 2 {
		t.Errorf("outs: %v, errCount: %v\n", ToJSON(outs), errCount)
		return
	}
	for k, out := range outs {
		if out.Name == "" || k == "" {
			t.Errorf("outs: %v, errCount: %v\n", ToJSON(outs), errCount)
			return
		}
	}
	fmt.Printf("outs: %v, errCount: %v\n", ToJSON(outs), errCount)
}
