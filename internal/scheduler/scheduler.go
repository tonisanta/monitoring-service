package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type Scheduler struct {
	ticker <-chan time.Time
}

func NewScheduler(ticker <-chan time.Time) *Scheduler {
	return &Scheduler{
		ticker: ticker,
	}
}

func (s *Scheduler) Run(ctx context.Context, fn func(context.Context)) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping scheduler")
			return
		case <-s.ticker:
			slog.Info("ticker has been triggered")
			select {
			case <-ctx.Done():
				slog.Info("ticker was triggered, but ctx to stop has higher priority")
				return
			default:
				fn(ctx)
			}
		}
	}
}
