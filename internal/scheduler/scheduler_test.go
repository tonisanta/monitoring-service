package scheduler_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"monitor-endpoint/internal/scheduler"
	"sync"
	"testing"
	"time"
)

func TestScheduler_Run(t *testing.T) {

	t.Run("should execute fn once for every tick", func(t *testing.T) {
		fakeTicker := make(chan time.Time)
		sched := scheduler.NewScheduler(fakeTicker)
		ctx, cancelFn := context.WithCancel(context.Background())

		spy := spyStruct{
			Done: make(chan struct{}),
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			sched.Run(ctx, spy.Run)
		}()

		numExecutions := 4
		go func() {
			for i := 0; i < numExecutions; i++ {
				fakeTicker <- time.Time{}
			}
		}()

		// wait for completion
		for i := 0; i < numExecutions; i++ {
			<-spy.Done
		}

		// scheduler should stop and counter should be equal to numExecutions
		cancelFn()
		wg.Wait()
		assert.Equal(t, numExecutions, spy.Counter)
	})

	t.Run("should stop immediately once ctx is cancelled", func(t *testing.T) {
		fakeTicker := make(chan time.Time, 1)
		doneChan := make(chan struct{}, 1)
		sched := scheduler.NewScheduler(fakeTicker)
		ctx, cancelFn := context.WithCancel(context.Background())
		spy := spyStruct{
			Done: doneChan,
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			sched.Run(ctx, spy.Run)
		}()

		// trigger an execution and wait for completion
		fakeTicker <- time.Time{}
		<-spy.Done
		assert.Equal(t, 1, spy.Counter)

		// tick is available, but ctx it's already cancelled
		cancelFn()
		fakeTicker <- time.Time{}

		// scheduler should stop with only 1 execution
		wg.Wait()
		assert.Equal(t, 1, spy.Counter)
	})
}

type spyStruct struct {
	Counter int
	Done    chan struct{}
}

func (s *spyStruct) Run(ctx context.Context) {
	s.Counter++
	s.Done <- struct{}{}
}
