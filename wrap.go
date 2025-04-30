package executors

import (
	"log/slog"
	"reflect"
)

var (
	_errType = reflect.TypeOf((*error)(nil)).Elem()
)

func WrapTaskG[Input Stateful, Output Stateful](fn TaskFuncG[Input, Output]) TaskFunc {
	return func(sin Stateful) (Stateful, error) {
		var in, err = Cast[Input](sin)
		if err != nil {
			return sin, err
		}
		return fn(in)
	}
}

func WrapTaskFunc(fn any) TaskFunc {
	if out, ok := fn.(TaskFunc); ok {
		return out
	}
	var (
		rerr    error
		fnRef   = reflect.ValueOf(fn)
		fnType  = fnRef.Type()
		errType = fnType.Out(1)
		inType  = fnType.In(0)
	)
	if fnRef.Kind() != reflect.Func {
		rerr = NewRuntimeErrorf("not a function, kind: %v", fnRef.Kind())
		goto end
	}
	if fnType.NumIn() != 1 {
		rerr = NewRuntimeErrorf("invalid function, NumIn: %v", fnType.NumIn())
		goto end
	}
	if fnType.NumOut() != 2 {
		rerr = NewRuntimeErrorf("invalid function, NumOut: %v", fnType.NumOut())
		goto end
	}
	if !errType.Implements(_errType) {
		rerr = NewRuntimeErrorf("invalid function, errType: %v not error", errType)
		goto end
	}
end:
	return func(sin Stateful) (Stateful, error) {
		if rerr != nil {
			return sin, rerr
		}
		var in = reflect.New(inType).Elem()
		err := reflectCastTo(reflect.ValueOf(&sin).Elem(), in)
		if err != nil {
			return sin, err
		}
		out := fnRef.Call([]reflect.Value{in})
		sout := out[0].Interface()
		errOut := out[1]
		if errOut.IsNil() {
			return sout, nil
		}
		return sout, errOut.Interface().(error)
	}
}

func WrapRemoveError(fn TaskFunc) TaskFunc {
	return func(state Stateful) (Stateful, error) {
		out, err := fn(state)
		if err != nil {
			slog.Error("", "error", err)
		}
		return out, nil
	}
}

func WrapWithCallback(fn TaskFunc, cb func(sout Stateful, err error)) TaskFunc {
	return func(sin Stateful) (Stateful, error) {
		out, err := fn(sin)
		cb(out, err)
		return out, err
	}
}
