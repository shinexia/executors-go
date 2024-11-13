package executors

import (
	"fmt"
	"testing"
)

func TestPostman(t *testing.T) {
	var out, expect int
	var prev = -1
	var consumer Consumer[[]int] = func(t []int) {
		if t[0] <= prev {
			panic(fmt.Sprintf("prev: %v, new: %v", prev, t[0]))
		}
		prev = t[0]
		out += Sum(t)
	}
	postman := NewPostman(consumer)
	N := 10000000
	for i := 0; i < N; i++ {
		postman.Post(i)
		expect += i
	}
	postman.Close()
	if out != expect {
		t.Errorf("invalid out: %v, expect: %v", out, expect)
		return
	}
}
