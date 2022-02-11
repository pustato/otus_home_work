package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	taskCh := make(chan Task)
	var errorCounter int64
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)

		go (func() {
			defer wg.Done()

			for t := range taskCh {
				if err := t(); err != nil {
					atomic.AddInt64(&errorCounter, 1)
				}
			}
		})()
	}

	for _, t := range tasks {
		if m > 0 && atomic.LoadInt64(&errorCounter) >= int64(m) {
			break
		}

		taskCh <- t
	}
	close(taskCh)

	wg.Wait()

	if m > 0 && errorCounter >= int64(m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}
