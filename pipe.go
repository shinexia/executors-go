package executors

import (
	"reflect"
)

const (
	_runtimeFieldStep = "__runtime_step"
)

type PipeStateful struct {
	Step  int `json:"__runtime_step"`
	Count int `json:"__runtime_count"`
	Data  any `json:"__runtime_data"`
}

type pipeRunner struct {
	tasks     []TaskFunc
	taskCount int
}

// Pipe prev task output will be used as next task's input, like unix's pipe
func Pipe(tasks ...any) TaskFunc {
	var taskCount = len(tasks)
	if taskCount == 0 {
		return func(sin Stateful) (Stateful, error) {
			return sin, nil
		}
	}
	var fnList = make([]TaskFunc, taskCount)
	for i, t := range tasks {
		fnList[i] = WrapTaskFunc(t)
	}
	if taskCount == 1 {
		return fnList[0]
	}
	return (&pipeRunner{
		tasks:     fnList,
		taskCount: taskCount,
	}).Run
}

func (r *pipeRunner) Run(sin Stateful) (Stateful, error) {
	var sinRef = reflect.ValueOf(&sin).Elem()
	var ctx PipeStateful
	if reflectContainsField(sinRef, _runtimeFieldStep) {
		err := reflectCastTo(sinRef, reflect.ValueOf(&ctx).Elem())
		if err != nil {
			return sin, err
		}
		if ctx.Count != r.taskCount {
			return sin, NewRuntimeErrorf("invalid count: %v, taskCount: %v", ctx.Count, r.taskCount)
		}
		if ctx.Step < 0 || ctx.Step >= r.taskCount {
			return sin, NewRuntimeErrorf("invalid step: %v, taskCount: %v", ctx.Step, r.taskCount)
		}
	} else {
		ctx.Count = r.taskCount
		ctx.Step = 0
		ctx.Data = sin
	}
	for ; ctx.Step < r.taskCount; ctx.Step++ {
		out, err := r.tasks[ctx.Step](ctx.Data)
		ctx.Data = out
		if err != nil {
			return ctx, err
		}
	}
	// succeed, remove context
	return ctx.Data, nil
}
