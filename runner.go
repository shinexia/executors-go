package executors

import (
	"time"
)

// taskRunner is a single task executor.
type taskRunner struct {
	queue    string
	spec     TaskSpec
	err      error
	runOpt   runOption
	state    TaskState
	delay    time.Duration
	fastFail chan int32
}

func Run(name string, fn TaskFunc, args Stateful, opts ...Option) (Stateful, error) {
	return RunSpec(TaskSpec{
		Name: name,
		Exec: fn,
		Args: args,
		Opts: opts,
	})
}

func RunSpec(spec TaskSpec, opts ...Option) (Stateful, error) {
	if spec.Exec == nil {
		return spec.Args, nil
	}
	tr := newTaskRunner("default", spec, newRunOption(spec.Opts, opts))
	tr.Run()
	return tr.state.Stateful, tr.err
}

func RunSpecList(specs []TaskSpec, opts ...Option) (Stateful, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	var sout = make([]Stateful, 0, len(specs))
	var errOut error
	for _, spec := range specs {
		out, err := RunSpec(spec, opts...)
		sout = append(sout, out)
		if err != nil {
			errOut = AppendError(errOut, err)
		}
	}
	return sout, errOut
}

func newTaskRunner(queue string, spec TaskSpec, ro runOption) *taskRunner {
	now := time.Now()
	tr := &taskRunner{
		queue:    queue,
		spec:     spec,
		runOpt:   ro,
		state:    NewTaskState(&spec, now),
		fastFail: make(chan int32, 1),
	}
	return tr
}

func (tr *taskRunner) Run() {
	tr.runOnce()
	if tr.delay > 0 {
		tr.slowLoop()
	}
	// let it crash if done() also panic
	tr.done()
}

func (tr *taskRunner) RunOnce() bool {
	tr.runOnce()
	if tr.delay > 0 {
		return false
	}
	// let it crash if done() also panic
	tr.done()
	return true
}

func (tr *taskRunner) RunTail() {
	tr.slowLoop()
	// let it crash if done() also panic
	tr.done()
}

func (tr *taskRunner) FastFail() {
	// maybe call after done, so don't close fastFail channel
	select {
	case tr.fastFail <- 1:
	default:
	}
}

func (tr *taskRunner) slowLoop() {
	// run once at least
	for tr.delay > 0 {
		tr.onRetry()
		select {
		case <-time.NewTimer(tr.delay).C:
			tr.runOnce()
		case <-tr.fastFail:
			tr.runOnce()
			return
		}
	}
}

func (tr *taskRunner) runOnce() {
	tr.doRun()
	tr.afterRun()
	if tr.err != nil {
		if !IsRuntimeError(tr.err) {
			delay := tr.retrable()
			if delay > 0 {
				select {
				case <-tr.fastFail:
					tr.delay = 0
					return
				default:
					tr.delay = delay
					return
				}
			}
		}
	}
	tr.delay = 0
}

func (tr *taskRunner) doRun() {
	defer func() {
		if err := recover(); err != nil {
			// find out exactly what the error was and set err
			switch x := err.(type) {
			case error:
				tr.err = NewRuntimeError(x)
			default:
				tr.err = NewRuntimeErrorf("%+v", err)
			}
		}
	}()
	tr.state.Stateful, tr.err = tr.spec.Exec(tr.state.Stateful)
}

func (tr *taskRunner) retrable() time.Duration {
	ro := &tr.runOpt
	executedCount := tr.state.ExecutedCount
	if ro.expiration <= 0 && executedCount >= ro.retryCount {
		return 0
	}
	var delay time.Duration
	if ro.backoff != nil {
		delay = ro.backoff(executedCount)
	} else {
		delay = DefaultBackoff(executedCount)
	}
	if ro.expiration > 0 {
		d2 := time.Until(tr.state.CreatedAt.Add(ro.expiration))
		if d2 <= 0 {
			return 0
		}
		delay = Min(delay, d2)
	}
	return delay
}

func (tr *taskRunner) afterRun() {
	tr.state.ExecutedCount++
	if tr.err != nil {
		tr.state.Error = tr.err.Error()
	} else {
		tr.state.Error = ""
	}
	tr.state.UpdatedAt = time.Now()

}

func (tr *taskRunner) onRetry() {
	tr.callback()
	for _, cb := range tr.runOpt.retryCallback {
		cb(&tr.state, tr.delay, tr.err)
	}
}

func (tr *taskRunner) callback() {
	for _, cb := range tr.runOpt.callback {
		cb(&tr.state, tr.err)
	}
}

func (tr *taskRunner) done() {
	tr.state.Finished = true
	tr.state.Success = tr.err == nil
	tr.callback()
}
