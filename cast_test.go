package executors

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestCastInt(t *testing.T) {
	cases := []struct {
		in  any
		out int
	}{
		{
			in:  int(1),
			out: int(1),
		},
		{
			in:  int8(1),
			out: int(1),
		},
		{
			in:  int16(1),
			out: int(1),
		},
		{
			in:  int32(1),
			out: int(1),
		},
		{
			in:  int64(1),
			out: int(1),
		},
		{
			in:  uint(1),
			out: int(1),
		},
		{
			in:  uint8(1),
			out: int(1),
		},
		{
			in:  uint16(1),
			out: int(1),
		},
		{
			in:  uint32(1),
			out: int(1),
		},
		{
			in:  uint64(1),
			out: int(1),
		},
		{
			in:  "1",
			out: int(1),
		},
	}
	for _, c := range cases {
		out, err := Cast[int](c.in)
		if err != nil {
			t.Errorf("out: %v, in: %v(%v), expect: %v, err: %v", out, reflect.TypeOf(c.in), c.in, c.out, err)
			continue
		}
		if out != c.out {
			t.Errorf("out: %v, in: %v(%v), expect: %v", out, reflect.TypeOf(c.in), c.in, c.out)
		}
	}
}

func TestInterface(t *testing.T) {
	var dst any
	var src = int(1)

	dstValue := reflect.ValueOf(&dst)
	fmt.Printf("dst type: %v, kind: %v, value: %v\n", dstValue.Type(), dstValue.Kind(), dstValue)

	err := CastTo(src, &dst)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if !reflect.DeepEqual(dst, src) {
		t.Errorf("dst: %v(%v), src: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(src), src)
		return
	}
}

func TestInterface2(t *testing.T) {
	type Test struct {
		Fail []string `json:"fail"`
	}
	var expect = Test{
		Fail: []string{"1", "2", "3"},
	}
	var src, err = DumpAndLoadTest(expect)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	dst, err := Cast[Test](src)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if !reflect.DeepEqual(dst, expect) {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
		return
	}
}

func TestTypeDef(t *testing.T) {
	type Number int

	var src int8 = 1

	dst, err := Cast[Number](src)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if dst != Number(src) {
		t.Errorf("dst: %v(%v), src: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(src), src)
		return
	}
}

func TestMapOp(t *testing.T) {
	var dst map[string]int
	x := reflect.ValueOf(&dst).Elem()
	x.Set(reflect.MakeMapWithSize(x.Type(), 0))
	dst["a"] = 1
}

func TestCastMap(t *testing.T) {
	var src = map[string]any{
		"1": 1,
		"2": "1",
		"3": true,
	}
	var expect = map[int]int{
		1: 1,
		2: 1,
		3: 1,
	}
	dst, err := Cast[map[int]int](src)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if !reflect.DeepEqual(dst, expect) {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
		return
	}
}

func TestArrayOp(t *testing.T) {
	var dst [3]int
	x := reflect.ValueOf(&dst).Elem()
	elmType := x.Type().Elem()
	x.Index(0).Set(reflect.Zero(elmType))
}

func TestCastArray(t *testing.T) {
	var src = []string{"1", "2", "3"}
	var expect = [3]int{1, 2, 3}
	var dst = [3]int{2, 4, 6}
	err := CastTo(src, &dst)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if !reflect.DeepEqual(dst, expect) {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
		return
	}
}

func TestCastArrayEmpty(t *testing.T) {
	var src = []string{}
	var expect = [3]int{0, 0, 0}
	var dst = [3]int{2, 4, 6}
	err := CastTo(src, &dst)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if !reflect.DeepEqual(dst, expect) {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
	}
}

func TestCastSlice(t *testing.T) {
	var src = []string{"1", "2", "3"}
	var expect = []int{1, 2, 3}
	dst, err := Cast[[]int](src)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if !reflect.DeepEqual(dst, expect) {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
		return
	}
}

func TestCastSliceEmpty(t *testing.T) {
	var src = []string{}
	var expect = []int{}
	var dst = []int{2, 4, 6}
	err := CastTo(src, &dst)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if !reflect.DeepEqual(dst, expect) {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
		return
	}
}

type Person struct {
	Name    string `json:"name,omitempty"`
	Age     int    `json:"age"`
	Ingored string `json:"-"`
	hidden  string
}

func TestStructOp(t *testing.T) {
	p := Person{}
	dst := reflect.ValueOf(&p).Elem()
	dstType := dst.Type()
	for i := 0; i < dstType.NumField(); i++ {
		field := dstType.Field(i)
		fmt.Printf("name: %v, tag: %v, IsExported: %v\n", field.Name, field.Tag.Get("json"), field.IsExported())
	}
	src := reflect.ValueOf(map[string]string{
		"name": "",
	})
	dst.Field(0).Set(src.MapIndex(reflect.ValueOf("name")))
	fmt.Printf("p: %v\n", ToJSON(p))

	if reflect.TypeOf(p) != reflect.TypeOf((*Person)(nil)).Elem() {
		t.Errorf("TypeOf is dynamic")
	} else {
		fmt.Println("TypeOf can use ==")
	}
}

func TestCastStruct(t *testing.T) {
	expect := Person{
		Name: "a",
		Age:  1,
	}
	data, err := json.Marshal(Person{
		Name:    expect.Name,
		Age:     expect.Age,
		Ingored: "ignored_msg",
		hidden:  "hidden_msg",
	})
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	var src any
	err = json.Unmarshal(data, &src)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	fmt.Printf("src: %v\n", src)
	dst, err := Cast[Person](src)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if !reflect.DeepEqual(dst, expect) {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
		return
	}
}

func TestPtrOp(t *testing.T) {
	var dst *int
	var src any = 1
	dstV := reflect.ValueOf(&dst).Elem()
	fmt.Printf("dst kind: %v, type: %v, isNil: %v\n", dstV.Kind(), dstV.Type(), reflectIsNil(dstV))
	srcV := reflect.ValueOf(&src).Elem()
	for (srcV.Kind() == reflect.Ptr || srcV.Kind() == reflect.Interface) && !srcV.IsNil() {
		srcV = srcV.Elem()
	}
	fmt.Printf("src kind: %v, type: %v, isNil: %v\n", srcV.Kind(), srcV.Type(), reflectIsNil(srcV))
	if srcV.CanAddr() {
		t.Errorf("interface cannot addr")
	}
}

func TestPtrInterface(t *testing.T) {
	var expect = 1
	var dst, err = Cast[*int](expect)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if *dst != expect {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
	}
	if dst == &expect {
		t.Errorf("pointer asigned not with copy")
	}
}

func TestPtr(t *testing.T) {
	var expect = 1
	var dst, err = Cast[*int](&expect)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if *dst != expect {
		t.Errorf("dst: %v(%v), expect: %v(%v)", reflect.TypeOf(dst), dst, reflect.TypeOf(expect), expect)
	}
	if dst != &expect {
		t.Errorf("pointer asigned not with copy")
	}
}

func TestCastMapContext(t *testing.T) {
	ctx := MapStateful[int, string]{
		runList: map[string]int{
			"0": 1,
			"1": 2,
			"2": 3,
		},
		Fail: map[string]any{
			"0": "x",
			"1": "1y",
		},
		SucceedList: []string{
			"x",
			"y",
			"z",
		},
	}
	out, err := DumpAndLoadTest(ctx)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	dst, err := Cast[MapStateful[int, string]](out)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	err = dst.SetOutput("0", "w")
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	fmt.Printf("dst: %v\n", ToJSON(dst))
}

func TestReflectContainsField(t *testing.T) {
	var ctx MapStateful[int, int]
	out := reflectContainsField(reflect.ValueOf(&ctx), _runtimeFieldFail)
	if !out {
		t.Errorf("MapContext should contains: %v", _runtimeFieldFail)
	}
}

func TestCastTime(t *testing.T) {
	src := time.Now().Format(time.RFC3339)
	dst, err := Cast[time.Time](src)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	fmt.Printf("dst: %v\n", ToJSON(dst))
}

func TestUnmarshalTime(t *testing.T) {
	src := time.Now().Format(time.RFC3339)
	fmt.Printf("src: %v\n", src)
	var dst time.Time
	err := json.Unmarshal([]byte("\""+src+"\""), &dst)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	fmt.Printf("dst: %v\n", ToJSON(dst))
}

func TestStructTypeComp(t *testing.T) {
	var src ParallelStateful
	var srcType = reflect.ValueOf(&src).Elem().Type()
	var dstType = reflect.TypeOf((*ParallelStateful)(nil)).Elem()
	if srcType != dstType {
		t.Errorf("dst type: %v, src type: %v", dstType, srcType)
	}
}
