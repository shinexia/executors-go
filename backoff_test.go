package executors

import (
	"fmt"
	"testing"
)

func TestExponentialBackoff(t *testing.T) {
	bo := DefaultBackoff
	for i := 0; i < 15; i++ {
		out := bo(i)
		fmt.Printf("i: %v, delay: %v\n", i, out)
	}
}
