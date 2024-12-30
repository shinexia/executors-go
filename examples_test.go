package executors_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/shinexia/executors-go"
)

var (
	_retryExpiration = 1 * time.Second
	_retryCallback   = func(state *executors.TaskState, delay time.Duration, err error) {
		fmt.Printf("task: %v, with retry after: %v, for err: %v\n", state.Name, delay, err)
	}
)

func TestGettingStarted(t *testing.T) {
	task := executors.Pipe(
		// split
		func(sin int) (any, error) {
			if rand.Intn(100) < 50 {
				// must return the original sin if error
				return sin, fmt.Errorf("split error")
			}
			out := make([]int, sin)
			for i := range sin {
				out[i] = i
			}
			return out, nil
		},
		// map
		executors.Map(func(sin int) (any, error) {
			if n := rand.Intn(100); n < 50 {
				// must return the original sin if error
				return sin, fmt.Errorf("map error")
			}
			return sin * 100, nil
		}),
		// reduce
		func(sin []int) (any, error) {
			if n := rand.Intn(100); n < 50 {
				// must return the original sin if error
				return sin, fmt.Errorf("reduce error")
			}
			return executors.Sum(sin), nil
		},
	)
	var snapshot []byte = nil
	var result int
	for {
		var sin any
		if len(snapshot) == 0 {
			sin = 10
		} else {
			// load state from snapshot
			state := &executors.TaskState{}
			_ = json.Unmarshal(snapshot, state)
			sin = state.Stateful
		}
		sout, err := executors.Run("test", task, sin, executors.WithCallback(func(state *executors.TaskState, err error) {
			// dump sout to a file or db
			snapshot, _ = json.Marshal(state)
		}))
		if err != nil {
			fmt.Printf("retring, sout: %v, err: %v\n", executors.ToJSON(sout), err)
			if executors.IsRuntimeError(err) {
				t.Fatalf("runtime error: %+v", err)
			}
			continue
		}
		result, _ = executors.Cast[int](sout)
		fmt.Printf("succeed, sout: %v\n", sout)
		break
	}
	if result != 4500 {
		t.Errorf("result: %v, expect: 4500", result)
	}
}

func TestDumpExecutor(t *testing.T) {
	executor := executors.NewExecutor(
		"dump-snapshot-executor",
		executors.WithCleanup(true),  // Clean up blocked tasks on new state
		executors.WithRunOnce(true),  // Execute tasks once in the main thread, avoiding goroutines
		executors.WithSkipPrev(true), // Skip all previous tasks if any exist
		executors.WithExpiration(_retryExpiration),
		executors.WithRetryCallback(_retryCallback),
	)
	defer executor.Close()
	snapshot := "snapshot"
	executor.Submit("test", func(sin any) (any, error) {
		_, err := fmt.Fprint(os.Stdout, snapshot)
		return sin, err
	}, nil)
}
