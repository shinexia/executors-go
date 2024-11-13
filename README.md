# Executors

Implement a stateful task executor, `Executor`, with the following features:

- Events are stateful and retryable, with the function signature: `func(sin any) (sout any, err error)`.
- Tasks can be orchestrated using `Pipe` (sequential), `Map` (data parallel), and `Parallel` (compute parallel) methods.
- For a multi-step task, if a step fails, the next retry will resume from the failed step, and previously successful steps will not be re-executed.
- Task states can be dumped and loaded, and upon reloading, execution resumes from the last failed step.
- Tasks automatically retry upon failure, but if new events arrive, the retrying tasks can be canceled. For example, when dumping a snapshot, if a new version of the snapshot is available, only the latest snapshot needs to be written, and previously blocked tasks can be canceled.

## Getting started

```go
import (
	"fmt"
	"math/rand"

	"github.com/shinexia/executors-go"
)

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
    var sin any = nil
    if len(snapshot) == 0 {
        sin = 10
    } else {
        // load state from snapshot
        state := &executors.TaskState{}
        json.Unmarshal(snapshot, state)
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
```

## More examples

See [examples_test.go](examples_test.go)
