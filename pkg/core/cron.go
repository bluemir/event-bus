package core

import (
	"context"
	"time"
)

func cron(ctx context.Context, f func(ctx context.Context) error, d time.Duration) func() error {
	return func() error {
		tick := time.NewTicker(1 * time.Minute)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				if err := f(ctx); err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
