package executors

import (
	"encoding/json"
	"fmt"
	"slices"
)

// number copied from golang.org/x/exp/constraints
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Min returns the minimum value among the given values.
func Min[T number](a T, xs ...T) T {
	for _, x := range xs {
		if x < a {
			a = x
		}
	}
	return a
}

// Sum returns the sum of the given values.
func Sum[T number](xs []T) T {
	var s T
	for _, x := range xs {
		s += x
	}
	return s
}

// Repeat returns a slice with the given value repeated n times.
func Repeat[T any](s T, n int) []T {
	out := make([]T, n)
	for i := 0; i < n; i++ {
		out[i] = s
	}
	return out
}

// RemoveAt removes elements at the specified indices from the slice.
func RemoveAt[T any](xs []T, is ...int) []T {
	n := len(xs)
	if n == 0 {
		return xs
	}
	in := len(is)
	switch in {
	case 0:
		return xs
	case 1:
		return removeAt(xs, is[0])
	}
	slices.Sort(is)
	var offset = 0
	var rdx = is[0]
	for i := 1; i < in; i++ {
		next := is[i]
		copy(xs[offset:], xs[rdx+1:next])
		offset += next - rdx - 1
		rdx = next
	}
	copy(xs[offset:], xs[rdx+1:])
	offset += n - rdx // - 1
	var zero T
	for ; offset < n; offset++ {
		xs[offset] = zero
	}
	return xs[:n-in]
}

// removeAt removes elements at the specified indices from the slice.
func removeAt[T any](xs []T, i int) []T {
	n := len(xs)
	if n == 0 || i < 0 || i >= n {
		return xs
	}
	var zero T
	if i == 0 || i == n-1 {
		xs[i] = zero
		if i == 0 {
			// first
			return xs[1:]
		}
		// last
		return xs[:i]
	}
	copy(xs[i:], xs[i+1:])
	xs[n-1] = zero
	return xs[:n-1]
}

// Remove removes the first occurrence of the given value from the slice.
func Remove[T comparable](xs []T, t T) []T {
	for i, x := range xs {
		if x == t {
			return RemoveAt(xs, i)
		}
	}
	return xs
}

// RunTest runs the given task function and returns the output and error count.
func RunTest(fn TaskFunc, ins Stateful) (outs Stateful, errCount int, rerr error) {
	for {
		s, err := fn(ins)
		if err == nil {
			outs = s
			return
		}
		if IsRuntimeError(err) {
			fmt.Printf("%+v\n", err)
			rerr = err
			return
		}
		if errCount == 0 {
			fmt.Printf("RunTest error ctx: %v\n", ToJSON(s))
		}
		errCount++
		if errCount%4 == 2 || errCount%4 == 3 {
			ins, err = DumpAndLoadTest(s)
			if err != nil {
				rerr = err
				return
			}
		} else {
			ins = s
		}
	}
}

// DumpAndLoadTest simulates the dump and load process.
func DumpAndLoadTest(s Stateful) (Stateful, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return s, NewRuntimeErrorf("json.Marsharl err: %v", err)
	}
	var out Stateful
	err = json.Unmarshal(data, &out)
	if err != nil {
		return s, NewRuntimeErrorf("json.Unmarshal err: %v", err)
	}
	return out, nil
}
