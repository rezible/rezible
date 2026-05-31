package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/sourcegraph/conc/pool"

	rez "github.com/rezible/rezible"
)

type appRunner struct {
	services []rez.LifecycleService
}

func (a *appRunner) start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		p := pool.New().WithErrors().WithContext(ctx).WithFirstError()
		for _, l := range a.services {
			slog.Info("Starting " + strings.TrimLeft(fmt.Sprintf("%T", l), "*"))
			p.Go(l.Start)
		}
		errChan <- p.Wait()
	}()
	slog.Info("=== Starting Services ===")
	var servicesErr error
	select {
	case <-ctx.Done():
		servicesErr = ctx.Err()
	case poolErr := <-errChan:
		servicesErr = poolErr
	}
	slog.Info("=== Stopping Services ===")
	if servicesErr != nil && !errors.Is(servicesErr, context.Canceled) {
		return fmt.Errorf("run services: %s", servicesErr.Error())
	}
	return nil
}
