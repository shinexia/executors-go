package executors

import (
	"log/slog"
	"sync"
)

type Consumer[T any] func(t T)

// Postman is a single-threaded task processor that combines blocked tasks.
type Postman[T any] interface {
	Post(t T)
	Close()
}

type defaultPostman[T any] struct {
	lock      sync.Mutex
	consumers []Consumer[[]T]

	closed   bool
	stopC    chan int32
	recvC    chan int32
	recvList []T

	doneWaiter sync.WaitGroup
}

func NewPostman[T any](consumers ...Consumer[[]T]) Postman[T] {
	pm := &defaultPostman[T]{
		consumers:  consumers,
		closed:     false,
		stopC:      make(chan int32, 1),
		recvC:      make(chan int32, 1),
		recvList:   nil,
		doneWaiter: sync.WaitGroup{},
	}
	pm.doneWaiter.Add(1)
	go pm.runLoop()
	return pm
}

func (pm *defaultPostman[T]) Post(t T) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	if pm.closed {
		slog.Error("post after closed")
		return
	}
	pm.recvList = append(pm.recvList, t)
	select {
	case pm.recvC <- 1:
	default:
	}
}

func (pm *defaultPostman[T]) Close() {
	pm.lock.Lock()
	if pm.closed {
		pm.lock.Unlock()
		return
	}
	pm.closed = true
	// unlock before wait
	pm.lock.Unlock()
	select {
	case pm.stopC <- 1:
	default:
	}
	pm.doneWaiter.Wait()
}

func (pm *defaultPostman[T]) runLoop() {
	defer pm.doneWaiter.Done()
	for {
		select {
		case <-pm.recvC:
			pm.consume()
		case <-pm.stopC:
			pm.consume()
			return
		}
	}
}

func (pm *defaultPostman[T]) consume() {
	// no locks in main loop
	recv := pm.fetch()
	if len(recv) > 0 && len(pm.consumers) > 0 {
		for _, cs := range pm.consumers {
			cs(recv)
		}
	}
}

func (pm *defaultPostman[T]) fetch() []T {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	out := pm.recvList
	pm.recvList = nil
	return out
}
