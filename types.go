package executors

import (
	"time"
)

// Stateful values must could be deserialized by json.Unmarshal
type Stateful = any
type TaskFuncG[Input Stateful, Output Stateful] func(sin Input) (Output, error)
type TaskFunc = func(sin Stateful) (Stateful, error)

// TaskCallback is a function type that is called when a task completes.
// It receives the task state and an error if one occurred.
type TaskCallback = func(state *TaskState, err error)
type RetryCallback = func(state *TaskState, delay time.Duration, err error)

// Executor is an interface that defines methods for submitting and managing tasks.
type Executor interface {
	// Submit adds a new task to the executor's queue.
	// 1. This method returns immediately.
	// 2. Multiple submissions are executed concurrently.
	// 3. WithCleanup waits for prior tasks to complete without affecting subsequent submissions.
	Submit(name string, fn TaskFunc, args Stateful, opts ...Option)

	// SubmitSpec is similar to Submit but takes a TaskSpec.
	SubmitSpec(spec TaskSpec, opts ...Option)

	// Cleanup ensures all tasks are executed at least once, no retries, and blocks until all tasks are completed.
	Cleanup()

	// Close performs cleanup and waits for all tasks to finish.
	Close()
}

// TaskSpec defines the specification for a task.
type TaskSpec struct {
	// Name is the name of the task.
	Name string
	// Exec is the function to be executed for the task.
	Exec TaskFunc
	// Args are the arguments to be passed to the task function.
	Args Stateful
	// Opts are the options for configuring the task execution.
	Opts []Option
}

// TaskState represents the state of a task.
type TaskState struct {
	// Name is the name of the task.
	Name string `json:"name"`
	// Finished indicates whether the task has completed.
	Finished bool `json:"finished"`
	// Success indicates whether the task was successful.
	Success bool `json:"success"`
	// ExecutedCount is the number of times the task has been executed.
	ExecutedCount int `json:"executed_count"`
	// Stateful is the internal state of the task.
	Stateful Stateful `json:"stateful"`
	// Error contains any error message if the task failed.
	Error string `json:"error,omitempty"`
	// CreatedAt is the time when the task was created.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt is the time when the task was last updated.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (ts *TaskState) IsEmpty() bool {
	return ts == nil || ts.Name == "" || ts.CreatedAt.IsZero() || ts.UpdatedAt.IsZero()
}

func NewTaskState(spec *TaskSpec, t time.Time) TaskState {
	return TaskState{
		Name:          spec.Name,
		Finished:      false,
		Success:       false,
		ExecutedCount: 0,
		Stateful:      spec.Args,
		Error:         "",
		CreatedAt:     t,
		UpdatedAt:     t,
	}
}

func NewTaskStateList(specs []TaskSpec, t time.Time) []TaskState {
	states := make([]TaskState, len(specs))
	for i := range specs {
		states[i] = NewTaskState(&specs[i], t)
	}
	return states
}

type FailCounter[T any] struct {
	Data      T   `json:"data"`
	FailCount int `json:"fail_count"`
}

func EmptyTask(sin Stateful) (Stateful, error) {
	return sin, nil
}

func GetTaskSpecsName(specs []TaskSpec) []string {
	names := make([]string, len(specs))
	for i := range specs {
		names[i] = specs[i].Name
	}
	return names
}
