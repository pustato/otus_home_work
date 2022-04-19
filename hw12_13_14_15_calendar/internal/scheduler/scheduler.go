package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
)

type Scheduler struct {
	s             *gocron.Scheduler
	ctx           context.Context
	cancelContext context.CancelFunc
}

func New(parent context.Context) *Scheduler {
	s := gocron.NewScheduler(time.UTC)
	s.TagsUnique()

	ctx, cancel := context.WithCancel(parent)

	return &Scheduler{
		s:             s,
		ctx:           ctx,
		cancelContext: cancel,
	}
}

func (s *Scheduler) AddTask(duration string, task Task) error {
	if d, err := time.ParseDuration(duration); err == nil {
		return s.addTaskWithDuration(d, task)
	}

	return s.addTaskWithCron(duration, task)
}

func (s *Scheduler) addTaskWithDuration(d time.Duration, task Task) error {
	if _, err := s.s.Every(d).Do(func() {
		_ = task(s.ctx)
	}); err != nil {
		return fmt.Errorf("scheduler add task: %w", err)
	}

	return nil
}

func (s *Scheduler) addTaskWithCron(cronExpression string, task Task) error {
	if _, err := s.s.Cron(cronExpression).Do(func() {
		_ = task(s.ctx)
	}); err != nil {
		return fmt.Errorf("scheduler add task: %w", err)
	}

	return nil
}

func (s *Scheduler) Start() {
	s.s.StartBlocking()
}

func (s *Scheduler) Stop() {
	s.cancelContext()
	s.s.Stop()
}
