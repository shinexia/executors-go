package executors

import (
	"fmt"
	"log/slog"
	"sync"
)

// queuedExecutor is an implementation of the Executor interface.
type queuedExecutor struct {
	lock        sync.RWMutex
	name        string
	defaultOpts []Option
	defaultRO   runOption

	waitQ taskQueue
	runQ  taskQueue

	stopC chan int32
	recvC chan int32

	closed bool
}

type taskQueue struct {
	lock    sync.Mutex
	tasks   []*taskRunner
	waiters []*sync.WaitGroup
}

func NewExecutor(name string, opts ...Option) Executor {
	qe := &queuedExecutor{
		name:        name,
		defaultOpts: opts,
		defaultRO:   newRunOption(opts),
		recvC:       make(chan int32, 1),
		stopC:       make(chan int32, 1),
		closed:      false,
	}
	if qe.name == "" {
		qe.name = fmt.Sprintf("%p", qe)
	}
	go qe.runLoop()
	return qe
}

// Submit adds a new task to the executor's queue.
// It returns immediately and allows concurrent execution of multiple submissions.
// WithCleanup waits for prior tasks to complete without affecting subsequent submissions.
func (qe *queuedExecutor) Submit(name string, fn TaskFunc, args Stateful, opts ...Option) {
	qe.SubmitSpec(TaskSpec{
		Name: name,
		Exec: fn,
		Args: args,
		Opts: opts,
	})
}

func (qe *queuedExecutor) SubmitSpec(spec TaskSpec, opts ...Option) {
	if spec.Exec == nil {
		return
	}
	qe.lock.RLock()
	defer qe.lock.RUnlock()
	if qe.closed {
		slog.Error("submit after closed", "queue", qe.name)
		return
	}
	ro := newRunOption(qe.defaultOpts, spec.Opts, opts)
	tr := newTaskRunner(qe.name, spec, ro)
	qe.waitQ.add(tr)
	select {
	case qe.recvC <- 1:
	default:
	}
}

// Cleanup ensures all tasks are executed at least once, no retries, and blocks until all tasks are completed.
func (qe *queuedExecutor) Cleanup() {
	qe.lock.Lock()
	defer qe.lock.Unlock()
	if qe.closed {
		slog.Error("cleanup after closed", "queue", qe.name)
		return
	}
	qe.cleanupAll()
}

// Close performs cleanup and waits for all tasks to finish.
func (qe *queuedExecutor) Close() {
	qe.lock.Lock()
	defer qe.lock.Unlock()
	if qe.closed {
		slog.Error("repeatly close", "queue", qe.name)
		return
	}
	qe.closed = true
	qe.cleanupAll()
	qe.stopC <- 1
}

func (qe *queuedExecutor) runLoop() {
	for {
		select {
		case <-qe.recvC:
			qe.consume()
		case <-qe.stopC:
			qe.consume()
			return
		}
	}
}

func (qe *queuedExecutor) consume() {
	tasks := qe.waitQ.getTasks()
	n := len(tasks)
	if n == 0 {
		return
	}
	for _, tr := range tasks {
		if tr.runOpt.cleanup {
			qe.cleanupRunQ()
		}
		qe.runQ.add(tr)
		qe.waitQ.removeTask(tr)
		if tr.runOpt.runOnce {
			if tr.RunOnce() {
				qe.runQ.removeTask(tr)
			} else {
				go qe.runTail(tr)
			}
		} else {
			go qe.runAndDone(tr)
		}
	}
}

func (qe *queuedExecutor) runAndDone(tr *taskRunner) {
	tr.Run()
	qe.runQ.removeTask(tr)
}

func (qe *queuedExecutor) runTail(tr *taskRunner) {
	tr.RunTail()
	qe.runQ.removeTask(tr)
}

func (qe *queuedExecutor) cleanupAll() {
	qe.waitQ.fastFail()
	qe.runQ.fastFail()
	qe.waitQ.cleanupQ()
	qe.runQ.cleanupQ()
}

func (qe *queuedExecutor) cleanupRunQ() {
	qe.runQ.fastFail()
	qe.runQ.cleanupQ()
}

// taskQueue
func (q *taskQueue) add(tr *taskRunner) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.waiters) > 0 {
		slog.Error("should not come here", "queue", tr.queue, "waiters", len(q.waiters), "task", tr.spec.Name)
	}
	q.tasks = append(q.tasks, tr)
}

func (q *taskQueue) getTasks() []*taskRunner {
	q.lock.Lock()
	defer q.lock.Unlock()
	var n = len(q.tasks)
	var i = n - 1
	for ; i >= 0; i-- {
		tr := q.tasks[i]
		if tr.runOpt.skipPrev {
			break
		}
	}
	if i < 1 {
		return q.tasks
	}
	for _, w := range q.waiters {
		for j := 0; j < i; j++ {
			w.Done()
		}
	}
	q.tasks = q.tasks[i:]
	return q.tasks
}

func (tq *taskQueue) fastFail() {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	for _, t := range tq.tasks {
		t.FastFail()
	}
}

func (tq *taskQueue) cleanupQ() {
	tq.lock.Lock()
	n := len(tq.tasks)
	if n == 0 {
		tq.lock.Unlock()
		return
	}
	w := &sync.WaitGroup{}
	w.Add(n)
	tq.waiters = append(tq.waiters, w)
	tq.lock.Unlock()
	w.Wait()
	tq.removeWaiter(w)
}

func (tq *taskQueue) removeTask(tr *taskRunner) {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	tq.tasks = Remove(tq.tasks, tr)
	for _, w := range tq.waiters {
		w.Done()
	}
}

func (q *taskQueue) removeWaiter(waiter *sync.WaitGroup) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.waiters = Remove(q.waiters, waiter)
}
