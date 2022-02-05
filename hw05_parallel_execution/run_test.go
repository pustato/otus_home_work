package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestWithoutTimout(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("m=0 means ignore errors", func(t *testing.T) {
		n, m, tasksCount := 11, 0, 103
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		waitCh := make(chan struct{})
		var err error

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				<-waitCh

				return errors.New("some error")
			})
		}

		lock := make(chan struct{})
		go (func() {
			defer close(lock)

			err = Run(tasks, n, m)
		})()

		require.Eventuallyf(t, func() bool {
			return int(atomic.LoadInt32(&runTasksCount)) == n
		}, time.Second, time.Microsecond, "workers is not parallel")

		close(waitCh)

		<-lock
		require.NoError(t, err, "errors were not ignored")
		require.Equal(t, int(runTasksCount), tasksCount, "not all tasks were completed")
	})

	t.Run("edge case with single long error task", func(t *testing.T) {
		n, m, tasksCount := 7, 1, 61
		tasks := make([]Task, 0, tasksCount)

		var successTasksCount, errorTasksCount int32
		var successTasksWg sync.WaitGroup

		tasks = append(tasks, func() error {
			atomic.AddInt32(&errorTasksCount, 1)
			successTasksWg.Wait()

			return errors.New("some error")
		})

		for i := 1; i < tasksCount; i++ {
			successTasksWg.Add(1)
			tasks = append(tasks, func() error {
				defer successTasksWg.Done()

				atomic.AddInt32(&successTasksCount, 1)

				return nil
			})
		}

		err := Run(tasks, n, m)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.Equal(t, int32(tasksCount-1), successTasksCount, "actual success tasks: %d", successTasksCount)
		require.Equal(t, int32(1), errorTasksCount, "actual error tasks: %d", errorTasksCount)
	})
}
