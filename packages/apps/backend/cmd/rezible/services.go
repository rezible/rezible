package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"

	rez "github.com/rezible/rezible"
)

func withConfig(i do.Injector, fn func(rez.Config) error) error {
	cfg, cfgErr := do.Invoke[rez.Config](i)
	if cfgErr != nil {
		return fmt.Errorf("invoke config: %w", cfgErr)
	}
	return fn(cfg)
}

func withMigrationService(i do.Injector, fn func(rez.MigrationService) error) error {
	ms, msErr := do.Invoke[rez.MigrationService](i)
	if msErr != nil {
		return fmt.Errorf("invoke migration service: %w", msErr)
	}
	return fn(ms)
}

func withAiEvaluationService(i do.Injector, fn func(rezai.EvalScenarioRunner) error) error {
	svc, svcErr := do.Invoke[rezai.EvalScenarioRunner](i)
	if svcErr != nil {
		return fmt.Errorf("invoke evaluation service: %w", svcErr)
	}
	return fn(svc)
}

func runLifecycleServices[Entrypoint rez.LifecycleService](ctx context.Context, i do.Injector) error {
	lifecycles, lifecyclesErr := getServiceLifecyclesFor[Entrypoint](i)
	if lifecyclesErr != nil {
		return fmt.Errorf("get service lifecycles: %w", lifecyclesErr)
	}

	return lifecycles.run(ctx)
}

type serviceLifecycles map[string]rez.ServiceLifecycle

func (ls serviceLifecycles) run(ctx context.Context) error {
	slog.Info("=== Starting Services ===")

	const lifecycleStopTimeout = 10 * time.Second

	group, runCtx := errgroup.WithContext(ctx)
	for name, lifecycle := range ls {
		slog.Info("Starting " + name)

		for _, runFn := range lifecycle.StartFns {
			group.Go(func() error {
				runErr := runFn(runCtx)
				if runErr != nil {
					if ctxErr := runCtx.Err(); ctxErr != nil && errors.Is(runErr, ctxErr) {
						return nil
					}
					return fmt.Errorf("%s: %w", name, runErr)
				}
				if runCtx.Err() == nil {
					return fmt.Errorf("%s stopped unexpectedly", name)
				}
				return nil
			})
		}

		if lifecycle.StopFn != nil {
			group.Go(func() error {
				<-runCtx.Done()
				stopCtx, cancel := context.WithTimeout(context.WithoutCancel(runCtx), lifecycleStopTimeout)
				defer cancel()

				if stopErr := lifecycle.StopFn(stopCtx); stopErr != nil {
					return fmt.Errorf("stop %s: %w", name, stopErr)
				}
				return nil
			})
		}
	}

	servicesErr := group.Wait()
	slog.Info("=== Services Stopped ===")

	if servicesErr != nil {
		return fmt.Errorf("run services: %w", servicesErr)
	}
	return nil
}

func shutdownInjector(ctx context.Context, i do.Injector) error {
	cancelCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	shutdown := i.ShutdownWithContext(cancelCtx)
	var shutdownErr error
	for sd, sErr := range shutdown.Errors {
		if !errors.Is(sErr, context.Canceled) {
			fmt.Printf("\n\t[%s] ERROR: %s\n", sd.Service, sErr.Error())
			shutdownErr = errors.Join(shutdownErr, sErr)
		}
	}
	return shutdownErr
}

func getServiceLifecyclesFor[Entrypoint rez.LifecycleService](i do.Injector) (serviceLifecycles, error) {
	b := newServiceBootstrapper(i)

	intgReg := do.MustInvoke[rez.IntegrationRegistry](i)
	eventProcessors := do.MustInvoke[rez.ProviderEventProcessorRegistry](i)
	b.addSetupHook("integrationPackages", setupHook[rez.IntegrationDefinition]{
		matchFn: func(d do.ServiceDescription) bool {
			return strings.Contains(d.Service, "internal/integrations")
		},
		serviceFn: func(pkg rez.IntegrationDefinition) error {
			if regErr := intgReg.Register(pkg); regErr != nil {
				return fmt.Errorf("failed to register integration package: %w", regErr)
			}
			if procPkg, isEventProcessor := pkg.(rez.ProviderEventProcessor); isEventProcessor {
				provider := pkg.Provider()
				if _, exists := eventProcessors[provider]; exists {
					return fmt.Errorf("failed to register event processor for provider %q: provider already registered", provider)
				}
				eventProcessors[provider] = procPkg
			}
			return nil
		},
	})

	services := make(serviceLifecycles)
	b.addInvokeHook("getServiceLifecycles", func(svcName string, ls rez.LifecycleService) error {
		if lifecycle := ls.Lifecycle(); lifecycle != nil {
			services[svcName] = *lifecycle
		}
		return nil
	})

	b.addInitHook("initJobs", func(r jobs.Registrar) error {
		return r.Init(do.MustInvoke[jobs.Definition](i))
	})

	return services, b.runFor[Entrypoint]()
}

type (
	serviceHook[S any] func(S) error
	invokeHook[S any]  func(string, S) error

	setupHook[S any] struct {
		matchFn   func(do.ServiceDescription) bool
		serviceFn serviceHook[S]
	}

	bootstrapper struct {
		i           do.Injector
		setupHooks  []setupHook[any]
		invokeHooks []invokeHook[any]
		initHooks   []func() error
	}
)

func newServiceBootstrapper(i do.Injector) *bootstrapper {
	return &bootstrapper{i: i}
}

func (b *bootstrapper) getServiceName(svc any) string {
	return strings.TrimLeft(fmt.Sprintf("%T", svc), "*")
}

func (b *bootstrapper) addSetupHook[S any](hookName string, hook setupHook[S]) {
	b.setupHooks = append(b.setupHooks, setupHook[any]{
		matchFn: hook.matchFn,
		serviceFn: func(svc any) error {
			if s, ok := svc.(S); ok {
				if hookErr := hook.serviceFn(s); hookErr != nil {
					return fmt.Errorf("[setup hook %s] %s: %w", hookName, b.getServiceName(svc), hookErr)
				}
			}
			return nil
		},
	})
}

func (b *bootstrapper) addInvokeHook[S any](hookName string, hookFn invokeHook[S]) {
	b.invokeHooks = append(b.invokeHooks, func(name string, svc any) error {
		if cbSvc, ok := svc.(S); ok {
			if fnErr := hookFn(name, cbSvc); fnErr != nil {
				return fmt.Errorf("[invoke hook %s] %s: %w", hookName, name, fnErr)
			}
		}
		return nil
	})
}

func (b *bootstrapper) addInitHook[S any](hookName string, hookFn func(S) error) {
	b.initHooks = append(b.initHooks, func() error {
		s, invokeErr := do.Invoke[S](b.i)
		if invokeErr != nil {
			return fmt.Errorf("[init hook %s]: %w", hookName, invokeErr)
		}
		if fnErr := hookFn(s); fnErr != nil {
			return fmt.Errorf("[init hook %s]: %w", hookName, fnErr)
		}
		return nil
	})
}

func (b *bootstrapper) runFor[Entrypoint any]() error {
	for _, desc := range b.i.ListProvidedServices() {
		for _, hook := range b.setupHooks {
			if !hook.matchFn(desc) {
				continue
			}
			s, invokeErr := do.InvokeNamed[any](b.i, desc.Service)
			if invokeErr != nil {
				return fmt.Errorf("failed to invoke %s: %w", desc.Service, invokeErr)
			}
			if hookErr := hook.serviceFn(s); hookErr != nil {
				return fmt.Errorf("setup: %w", hookErr)
			}
		}
	}

	// invoke entrypoint service to invoke required service dependencies
	if entrySvc, srvErr := do.Invoke[Entrypoint](b.i); srvErr != nil {
		return fmt.Errorf("invoke entrypoint %T: %w", entrySvc, srvErr)
	}

	for _, hookFn := range b.initHooks {
		if hookErr := hookFn(); hookErr != nil {
			return fmt.Errorf("bootstrap: %w", hookErr)
		}
	}

	for _, desc := range b.i.ListInvokedServices() {
		svc, invErr := do.InvokeNamed[any](b.i, desc.Service)
		if invErr != nil {
			return fmt.Errorf("failed to invoke %s: %w", desc.Service, invErr)
		}
		for _, hookFn := range b.invokeHooks {
			if hookErr := hookFn(b.getServiceName(svc), svc); hookErr != nil {
				return fmt.Errorf("bootstrap: %w", hookErr)
			}
		}
	}

	return nil
}
