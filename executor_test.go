package executors

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecutorSucceed(t *testing.T) {
	fn := WrapTaskG(func(sin int) (int, error) {
		return sin + 100, nil
	})
	var out []interface{}
	cb := func(state *TaskState, err error) {
		out = append(out, state.Stateful)
	}
	qe := NewExecutor("test", WithCleanup(true), WithRetryCount(5), WithCallback(cb))
	in := []int{1, 2, 3}
	expect := []int{101, 102, 103}
	for i, s := range in {
		qe.Submit("test"+strconv.Itoa(i), fn, s)
	}
	qe.Close()
	cout, err := Cast[[]int](out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(cout) != len(expect) {
		t.Errorf("invalid out: %v, in: %v, expect: %v", cout, in, expect)
		return
	}
	for i, o := range cout {
		if o != expect[i] {
			t.Errorf("invalid out: %v, i: %v, in: %v, expect: %v", cout, i, in, expect)
			return
		}
	}
	fmt.Printf("cout: %v\n", cout)
}

func TestExecutorRetry(t *testing.T) {
	fn := WrapTaskG(func(sin int) (int, error) {
		if sin < 3 {
			return sin + 1, fmt.Errorf("inject_error")
		}
		return sin + 100, nil
	})
	wg := sync.WaitGroup{}
	var out []interface{}
	cb := func(state *TaskState, err error) {
		if state.Finished {
			out = append(out, state.Stateful)
			wg.Done()
		}
	}
	qe := NewExecutor("test",
		WithCleanup(true),
		WithRetryCount(10),
		WithCallback(cb),
		WithBackoff(FixedIntervalBackoff(100*time.Millisecond)),
	)
	in := []int{1, 2, 3}
	expect := []int{103, 103, 103}
	for i, s := range in {
		wg.Add(1)
		qe.Submit("test"+strconv.Itoa(i), fn, s)
		wg.Wait()
	}
	qe.Close()
	cout, err := Cast[[]int](out)
	if err != nil {
		t.Error(err)
		return
	}
	if len(cout) != len(expect) {
		t.Errorf("invalid out: %v, in: %v, expect: %v", cout, in, expect)
		return
	}
	for i, o := range cout {
		if o != expect[i] {
			t.Errorf("invalid out: %v, i: %v, in: %v, expect: %v", cout, i, in, expect)
			return
		}
	}
	fmt.Printf("cout: %v\n", cout)
}

func TestExecutorGetTasks(t *testing.T) {
	cases := []struct {
		in     []bool
		expect int
	}{
		{
			in:     []bool{false, false, false, false},
			expect: 4,
		},
		{
			in:     []bool{true, false, false, false},
			expect: 4,
		},
		{
			in:     []bool{true, true, false, false},
			expect: 3,
		},
		{
			in:     []bool{false, true, false, false},
			expect: 3,
		},
		{
			in:     []bool{false, true, true, false},
			expect: 2,
		},
		{
			in:     []bool{false, true, true, true},
			expect: 1,
		},
		{
			in:     []bool{true, true, true, true},
			expect: 1,
		},
	}
	for _, c := range cases {
		q := &taskQueue{}
		for i, b := range c.in {
			q.tasks = append(q.tasks, newTaskRunner("", TaskSpec{
				Name: strconv.Itoa(i),
			}, newRunOption([]Option{WithSkipPrev(b)})))
		}
		wg := &sync.WaitGroup{}
		wg.Add(len(c.in) - c.expect)
		q.waiters = append(q.waiters, wg)
		tasks := q.getTasks()
		wg.Wait()
		out := make([]int, len(tasks))
		for i, t := range tasks {
			out[i], _ = Cast[int](t.spec.Name)
		}
		if len(out) != c.expect {
			t.Errorf("out: %v, in: %v, expect: %v", out, c.in, c.expect)
			return
		}
		offset := len(c.in) - len(out)
		for i := offset; i < len(c.in); i++ {
			if out[i-offset] != i {
				t.Errorf("out: %v, i: %v, in: %v, expect: %v", out, i, c.in, c.expect)
				return
			}
		}
	}
}

func testCheckOrder(qe Executor, N int64, t *testing.T) {
	out := int64(0)
	expect := int64(0)
	prev := int64(-1)
	fn := func(sin any) (any, error) {
		return sin.(int64), nil
	}
	cb := func(state *TaskState, err error) {
		s := state.Stateful.(int64)
		if s <= prev {
			panic(fmt.Sprintf("prev: %v, s: %v", prev, s))
		}
		prev = s
		out += s
	}
	for i := int64(0); i < N; i++ {
		qe.Submit("test", fn, i, WithCallback(cb))
		expect += i
	}
	qe.Close()
	if out != expect {
		t.Errorf("out: %v, N: %v, expect: %v", out, N, expect)
	}
}

func TestExecutorOrder1(t *testing.T) {
	testCheckOrder(NewExecutor("test", WithCleanup(true)), 1000000, t)
}

func TestExecutorOrder2(t *testing.T) {
	testCheckOrder(NewExecutor("test", WithCleanup(true), WithRunOnce(true)), 1000000, t)
}

func testCheckParallel(qe Executor, N int64, t *testing.T) {
	out := int64(0)
	expect := int64(0)
	fn := func(sin any) (any, error) {
		return sin, nil
	}
	cb := func(state *TaskState, err error) {
		atomic.AddInt64(&out, state.Stateful.(int64))
	}
	for i := int64(0); i < N; i++ {
		qe.Submit("test", fn, i, WithCallback(cb))
		expect += i
	}
	qe.Close()
	if out != expect {
		t.Errorf("out: %v, N: %v, expect: %v", out, N, expect)
	}
}

func TestExecutorParallel1(t *testing.T) {
	testCheckParallel(NewExecutor("test"), 1000000, t)
}

func TestExecutorParallel2(t *testing.T) {
	testCheckParallel(NewExecutor("test", WithRunOnce(true)), 1000000, t)
}

func testSkipPrev(qe Executor, N int64, t *testing.T) {
	out := int64(0)
	expect := int64(0)
	fn := func(sin any) (any, error) {
		return sin, nil
	}
	cb := func(state *TaskState, err error) {
		atomic.AddInt64(&out, state.Stateful.(int64))
	}
	for i := int64(0); i < N; i++ {
		qe.Submit("test", fn, i, WithCallback(cb))
		expect += i
	}
	qe.Close()
	if out >= expect {
		t.Errorf("out: %v, N: %v, expect: %v", out, N, expect)
	}
}

func TestExecutorSkipPrev1(t *testing.T) {
	testSkipPrev(NewExecutor("test", WithSkipPrev(true)), 1000000, t)
}

func TestExecutorSkipPrev2(t *testing.T) {
	testSkipPrev(NewExecutor("test", WithSkipPrev(true), WithRunOnce(true)), 1000000, t)
}

func TestExecutorSkipPrev3(t *testing.T) {
	qe := NewExecutor("test",
		WithCleanup(true),
		WithSkipPrev(true),
		WithRunOnce(true),
		WithRetryCount(100),
		WithBackoff(FixedIntervalBackoff(100*time.Millisecond)),
	)
	Base := int64(100)
	N := int64(10000)
	wgs := make([]*sync.WaitGroup, N/Base)
	for i := range wgs {
		w := &sync.WaitGroup{}
		w.Add(1)
		wgs[i] = w
	}
	err1 := fmt.Errorf("test_error")
	fn := WrapTaskG(func(sin []int64) (any, error) {
		if sin[1]%Base == 0 {
			wgs[sin[0]/Base].Done()
		}
		sin[1] += 1
		if sin[1] >= 3 {
			fmt.Printf("sin: %v tried too many", sin)
			os.Exit(1)
		}
		return sin, err1
	})
	for i := int64(0); i < N; i++ {
		qe.Submit("test", fn, []int64{i, 0})
		if i%Base == 0 {
			wgs[i/Base].Wait()
		}
	}
	qe.Close()
}

func testsum(ni any) (any, error) {
	var s float64
	var n = ni.(int)
	for i := 0; i < n; i++ {
		s += math.Pow(2, float64(i))
	}
	return s, nil
}

const _testsumN = 0

func BenchmarkExecutorCleanup0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = testsum(_testsumN)
	}
}

func BenchmarkExecutorCleanup1(b *testing.B) {
	qe := NewExecutor("test", WithCleanup(true))
	for i := 0; i < b.N; i++ {
		qe.Cleanup()
		_, _ = testsum(_testsumN)
	}
	qe.Close()
}

func BenchmarkExecutorCleanup2(b *testing.B) {
	qe := NewExecutor("test", WithCleanup(true), WithRunOnce(true))
	for i := 0; i < b.N; i++ {
		qe.Submit("test", testsum, _testsumN)
	}
	qe.Close()
}

func BenchmarkExecutorCleanup3(b *testing.B) {
	qe := NewExecutor("test", WithCleanup(true))
	for i := 0; i < b.N; i++ {
		qe.Submit("test", testsum, _testsumN)
	}
	qe.Close()
}
