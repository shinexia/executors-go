package executors

import (
	"reflect"
	"sync"
)

type ParallelStateful struct {
	runList map[int]any `json:"-"`
	Succeed []any       `json:"__runtime_succeed"`
	Fail    []int       `json:"__runtime_fail"`
}

type parallelRunner struct {
	tasks     []TaskFunc
	taskCount int
}

// Parallel run the same data on different tasks; this is task parallel
func Parallel(tasks ...any) TaskFunc {
	if len(tasks) == 0 {
		return nil
	}
	var taskCount = len(tasks)
	var fnList = make([]TaskFunc, taskCount)
	for i, fn := range tasks {
		fnList[i] = WrapTaskFunc(fn)
	}
	return (&parallelRunner{
		tasks:     fnList,
		taskCount: taskCount,
	}).Run
}

func (r *parallelRunner) Run(sin Stateful) (Stateful, error) {
	var sinRef = reflect.ValueOf(sin)
	var ctx ParallelStateful
	if reflectContainsField(sinRef, _runtimeFieldFail) {
		err := reflectCastTo(sinRef, reflect.ValueOf(&ctx))
		if err != nil {
			return sin, err
		}
		if len(ctx.Fail) == 0 {
			return sin, nil
		}
		if len(ctx.Succeed) == 0 {
			return sin, NewRuntimeErrorf("convert ParallelStateful fail, sin %v(%v), both empty", sinRef.Type(), sin)
		}
		ctx.runList = make(map[int]any, len(ctx.Fail))
		for _, k := range ctx.Fail {
			if k < 0 || k > len(ctx.Succeed) {
				return sin, NewRuntimeErrorf("invalid k: %d, Succeed: %d", k, len(ctx.Succeed))
			}
			ctx.runList[k] = ctx.Succeed[k]
		}
		ctx.Fail = nil
	} else {
		ctx.runList = make(map[int]any, r.taskCount)
		for i := 0; i < r.taskCount; i++ {
			ctx.runList[i] = sin
		}
		ctx.Succeed = make([]any, r.taskCount)
	}
	var errOut error
	var lock sync.Mutex
	var wg = sync.WaitGroup{}
	wg.Add(len(ctx.runList))
	for k1, s1 := range ctx.runList {
		go func(k int, s any) {
			defer wg.Done()
			out, err := r.tasks[k](s)
			lock.Lock()
			ctx.Succeed[k] = out
			if err != nil {
				ctx.Fail = append(ctx.Fail, k)
				errOut = AppendError(errOut, err)
			}
			lock.Unlock()
		}(k1, s1)
	}
	wg.Wait()
	if errOut != nil {
		return ctx, errOut
	}
	return ctx.Succeed, nil
}
