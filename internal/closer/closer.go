package closer

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

type closeFn struct {
	name string
	fn   func(context.Context) error
}

type closer struct {
	mu    sync.Mutex
	once  sync.Once
	funcs []closeFn
}

var globalCloser = &closer{}

func Add(name string, fn func(context.Context) error) {
	globalCloser.add(name, fn)
}

func CloseAll(ctx context.Context) error {
	return globalCloser.closeAll(ctx)
}

func (c *closer) add(name string, fn func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, closeFn{name: name, fn: fn})
}

func (c *closer) closeAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			return
		}

		slog.Info("starting graceful shutdown", slog.Int("count", len(funcs)))

		var errs []error

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]

			start := time.Now()
			slog.Info("closing resource", slog.String("name", f.name))
			resourceCount := i + 1
			resourceTimeout := getResourceTimeout(ctx, resourceCount)
			resourceCtx, resourceCancel := context.WithTimeout(ctx, resourceTimeout)

			if err := f.fn(resourceCtx); err != nil {
				slog.Error("close resource",
					slog.String("name", f.name),
					slog.Any("error", err),
					slog.Duration("duration", time.Since(start)),
				)

				errs = append(errs, err)
			} else {
				slog.Info("resource closed",
					slog.String("name", f.name),
					slog.Duration("duration", time.Since(start)))
			}
			resourceCancel()
		}

		slog.Info("graceful shutdown completed")

		result = errors.Join(errs...)
	})

	return result
}

func getResourceTimeout(ctx context.Context, resourceCount int) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0
	}
	timeout := time.Until(deadline)

	if timeout <= 0 {
		return 0
	}

	return timeout / time.Duration(resourceCount)
}
