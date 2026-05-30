package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rezible/rezible/internal/http"
	"github.com/samber/do/v2"
	"github.com/sourcegraph/conc/pool"
)

func runServer(ctx context.Context, i do.Injector) error {
	// invoke http server
	_ = do.MustInvoke[*http.Server](i)

	fmt.Println("=== Starting Services ===")
	serviceErr := runServices(ctx, i)
	fmt.Println("=== Stopping Services ===")

	if serviceErr != nil && !errors.Is(serviceErr, context.Canceled) {
		return fmt.Errorf("start services: %s", serviceErr.Error())
	}
	return nil
}

type startable interface {
	Start(context.Context) error
}

func runServices(ctx context.Context, i do.Injector) error {
	var services []startable
	for _, desc := range i.ListInvokedServices() {
		if svc, ok := do.MustInvokeNamed[any](i, desc.Service).(startable); ok {
			services = append(services, svc)
		}
	}

	errChan := make(chan error)
	p := pool.New().WithErrors().WithContext(ctx)
	go func() {
		for _, l := range services {
			slog.Info("Starting " + strings.TrimLeft(fmt.Sprintf("%T", l), "*"))
			p.Go(l.Start)
		}
		errChan <- p.Wait()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case poolErr := <-errChan:
		return poolErr
	}
}
