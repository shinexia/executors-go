package executors

import (
	"reflect"
	"strconv"
	"sync"
)

var (
	_runtimeFieldFail = "__runtime_fail"
)

type MapStateful[Input Stateful, Output Stateful] struct {
	runList     map[string]Input  `json:"-"`
	Fail        map[string]any    `json:"__runtime_fail"`
	SucceedList []Output          `json:"__runtime_list,omitempty"`
	SucceedMap  map[string]Output `json:"__runtime_map,omitempty"`
}

type mapRunner[Input Stateful, Output Stateful] struct {
	task TaskFuncG[Input, Output]
}

// Map run elements of map or slices on the same task; this is data parallel
func Map[Input Stateful, Output Stateful](task TaskFuncG[Input, Output]) TaskFunc {
	return (&mapRunner[Input, Output]{
		task: task,
	}).Run
}

func (r *mapRunner[Input, Output]) Run(sin Stateful) (Stateful, error) {
	if sin == nil {
		return sin, nil
	}
	var sinRef = reflect.ValueOf(sin)
	var ctx MapStateful[Input, Output]
	if reflectContainsField(sinRef, _runtimeFieldFail) {
		err := reflectCastTo(sinRef, reflect.ValueOf(&ctx))
		if err != nil {
			return sin, err
		}
		if len(ctx.Fail) == 0 {
			return sin, nil
		}
		err = CastTo(ctx.Fail, &ctx.runList)
		if err != nil {
			return sin, err
		}
		ctx.Fail = make(map[string]any)
		if len(ctx.SucceedMap) == 0 {
			ctx.SucceedMap = make(map[string]Output)
		}
	} else {
		switch sinRef.Kind() {
		case reflect.Array, reflect.Slice:
			var runList []Input
			err := reflectCastTo(sinRef, reflect.ValueOf(&runList))
			if err != nil {
				return sin, err
			}
			if len(runList) == 0 {
				return sin, nil
			}
			ctx.runList = make(map[string]Input, len(runList))
			for i := range runList {
				ctx.runList[strconv.Itoa(i)] = runList[i]
			}
			ctx.SucceedList = make([]Output, len(runList))
			ctx.Fail = make(map[string]any)
		case reflect.Map:
			var runList map[string]Input
			err := reflectCastTo(sinRef, reflect.ValueOf(&runList))
			if err != nil {
				return sin, err
			}
			if len(runList) == 0 {
				return sin, nil
			}
			ctx.runList = runList
			// now is empty
			ctx.SucceedMap = make(map[string]Output, len(runList))
			ctx.Fail = make(map[string]any)
		default:
			return sin, NewRuntimeErrorf("convert MapStateful fail, sin %v(%v)", sinRef.Type(), sin)
		}
	}
	var errOut error
	var lock sync.Mutex
	var wg = sync.WaitGroup{}
	wg.Add(len(ctx.runList))
	for k1, s1 := range ctx.runList {
		go func(k string, s Input) {
			defer wg.Done()
			out, err := r.task(s)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				ctx.Fail[k] = out
				errOut = AppendError(errOut, err)
			} else {
				err = ctx.SetOutput(k, out)
				if err != nil {
					errOut = AppendError(errOut, err)
					return
				}
			}
		}(k1, s1)
	}
	wg.Wait()
	if errOut != nil {
		return ctx, errOut
	}
	return ctx.Succeed()
}

func (ctx *MapStateful[Input, Output]) SetOutput(k string, v Output) error {
	if len(ctx.SucceedList) > 0 {
		i, err := strconv.Atoi(k)
		if err != nil {
			return NewRuntimeErrorf("invalid k: %s", k)
		}
		if i < 0 || i >= len(ctx.SucceedList) {
			return NewRuntimeErrorf("invalid k: %s, SucceedList: %d", k, len(ctx.SucceedList))
		}
		ctx.SucceedList[i] = v
	} else {
		ctx.SucceedMap[k] = v
	}
	return nil
}

func (ctx *MapStateful[Input, Output]) Succeed() (Stateful, error) {
	if len(ctx.SucceedList) > 0 {
		return ctx.SucceedList, nil
	} else {
		return ctx.SucceedMap, nil
	}
}
