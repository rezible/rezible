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
	"github.com/riverqueue/river"
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

	intgReg := do.MustInvoke[rez.IntegrationPackageRegistry](i)
	eventProcessors := do.MustInvoke[rez.ProviderEventProcessorRegistry](i)
	b.addSetupHook("integrationPackages", setupHook[rez.IntegrationPackage]{
		matchFn: func(d do.ServiceDescription) bool {
			return strings.Contains(d.Service, "internal/integrations")
		},
		serviceFn: func(pkg rez.IntegrationPackage) error {
			if regErr := intgReg.RegisterPackage(pkg); regErr != nil {
				return fmt.Errorf("failed to register integration package: %w", regErr)
			}
			if procPkg, isEventProcessor := pkg.(rez.ProviderEventProcessor); isEventProcessor {
				eventProcessors[pkg.Name()] = procPkg
			}
			return nil
		},
	})

	jobsRegistry := do.MustInvoke[*jobs.Registry](i)

	b.addInvokeHook("registerJobs", func(svcName string, r jobs.JobRegistrar) error {
		return r.RegisterJobs(jobsRegistry)
	})

	b.addInvokeHook("registerWorkers", func(svcName string, r jobs.WorkerRegistrar) error {
		return r.RegisterWorkers(jobsRegistry)
	})

	services := make(serviceLifecycles)
	b.addInvokeHook("getServiceLifecycles", func(svcName string, ls rez.LifecycleService) error {
		if lifecycle := ls.Lifecycle(); lifecycle != nil {
			services[svcName] = *lifecycle
		}
		return nil
	})

	finalizers := map[string]finalizer{}
	b.addInvokeHook("addFinalizers", func(svcName string, f finalizer) error {
		finalizers[svcName] = f
		return nil
	})

	b.addRunHook("addWorkers", func() error {
		r := do.MustInvoke[jobs.WorkerRegistrar](i)
		return r.RegisterWorkers(jobsRegistry)
	})

	b.addRunHook("finalizers", func() error {
		for svcName, f := range finalizers {
			if finErr := f.Finalize(); finErr != nil {
				return fmt.Errorf("finalize %s: %w", svcName, finErr)
			}
		}
		return nil
	})

	return services, b.runFor[Entrypoint]()
}

type finalizer interface {
	Finalize() error
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
		runHooks    []func() error
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

func (b *bootstrapper) addRunHook(hookName string, hookFn func() error) {
	b.runHooks = append(b.runHooks, func() error {
		if fnErr := hookFn(); fnErr != nil {
			return fmt.Errorf("[run hook %s]: %w", hookName, fnErr)
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

	for _, hookFn := range b.runHooks {
		if hookErr := hookFn(); hookErr != nil {
			return fmt.Errorf("bootstrap: %w", hookErr)
		}
	}

	return nil
}

type defaultJobWorkers struct {
	i do.Injector
}

func newDefaultWorkerRegistrar(i do.Injector) *defaultJobWorkers {
	return &defaultJobWorkers{i: i}
}

func (r *defaultJobWorkers) RegisterWorkers(reg *jobs.Registry) error {
	return errors.Join(
		r.register[jobs.StartAgentSession](reg),
		r.register[jobs.InvokeAgentTurn](reg),
		r.register[jobs.SyncIntegrationSourceEvents](reg),
	)
}

func (r *defaultJobWorkers) register[A river.JobArgs](reg *jobs.Registry) error {
	kind := (*new(A)).Kind()
	slog.Debug("registered job worker", "kind", kind)

	worker, workerErr := do.Invoke[jobs.Worker[A]](r.i)
	if workerErr != nil {
		return fmt.Errorf("invoke worker for %q: %w", kind, workerErr)
	}
	if registerErr := reg.AddWorker(worker); registerErr != nil {
		return fmt.Errorf("register worker for %q: %w", kind, registerErr)
	}
	return nil
}
