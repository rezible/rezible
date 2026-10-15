package main

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"text/tabwriter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/evals"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/samber/do/v2"
)

// agentEvaluationProfile prepares the application for agent evaluation scenarios, on a test database: the
// evaluation integration and its event processor beside the given integrations, a signal service that lets
// setup events skip detection, and neither periodic jobs nor a model judge, since the scenario decides what the
// agent faces.
func agentEvaluationProfile(definitions Provider[[]rez.IntegrationDefinition]) Package {
	return do.Package(
		testDatabaseProfile,
		override(func(i do.Injector) ([]rez.IntegrationDefinition, error) {
			base, baseErr := definitions(i)
			if baseErr != nil {
				return nil, baseErr
			}
			return append(base, evals.NewIntegration(do.MustInvoke[rez.Clock](i))), nil
		}),
		override(func(i do.Injector) (rez.ProviderEventProcessorRegistry, error) {
			registry, registryErr := provideProviderEventProcessors(i)
			if registryErr != nil {
				return nil, registryErr
			}
			registry[evals.EvaluationProviderName] = evals.EventProcessor{}
			return registry, nil
		}),
		override(func(i do.Injector) (rez.SituationSignalService, error) {
			return evals.NewSetupSignals(db.NewSituationSignalService(do.MustInvoke[rez.JobService](i))), nil
		}),
		override(func(do.Injector) ([]*jobs.PeriodicJob, error) {
			return nil, nil
		}),
		override(func(do.Injector) (rezai.JudgeSituationCandidateWorkflow, error) {
			return nil, nil
		}),
	)
}

func printEvalTargets(w io.Writer, app *Application) error {
	return app.With(func(workflows []evals.WorkflowTarget) error {
		targets, scenariosErr := evals.ScenarioTargets()
		if scenariosErr != nil {
			return scenariosErr
		}
		for _, workflow := range workflows {
			targets = append(targets, workflow.Target())
		}
		table := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
		for _, target := range targets {
			_, _ = fmt.Fprintf(table, "%s\t%s\t%s\n", target.Name, target.Kind, target.Description)
		}
		return table.Flush()
	})
}

type evalRunner struct {
	target string
	runs   int
	outDir string
}

func newEvalRunner(target string, runs int, outDir string) *evalRunner {
	if runs < 1 {
		runs = 1
	}
	if outDir == "" {
		outDir = filepath.Join(".evals", target+"-"+time.Now().Format("20060102T150405"))
	}
	return &evalRunner{target: target, runs: runs, outDir: outDir}
}

func (r *evalRunner) run(ctx context.Context, w io.Writer) error {
	runFn := r.runWorkflowDataset
	var profiles []Package
	if evals.HasScenario(r.target) {
		runFn = r.runAgentScenario
		profiles = []Package{agentEvaluationProfile(provideIntegrationDefinitions)}
	}
	return runApplication(ctx, func(ctx context.Context, app *Application) error {
		result, runErr := runFn(ctx, app)
		if runErr != nil || result == nil {
			return fmt.Errorf("run: %w", runErr)
		}
		summary, reportErr := evals.Report(*result, r.outDir, w)
		if reportErr != nil {
			return reportErr
		}
		if summary.Errors > 0 {
			return fmt.Errorf("%d of %d runs errored", summary.Errors, summary.Runs)
		}
		return nil
	}, profiles...)
}

func (r *evalRunner) runAgentScenario(ctx context.Context, app *Application) (*evals.RunResult, error) {
	var result *evals.RunResult
	return result, app.runLifecycleFunc(ctx, func(ctx context.Context) error {
		return app.With(func(agents *evals.AgentEvalService) error {
			var runErr error
			result, runErr = agents.Run(ctx, r.target, r.runs)
			return runErr
		})
	})
}

func (r *evalRunner) runWorkflowDataset(ctx context.Context, app *Application) (*evals.RunResult, error) {
	var result *evals.RunResult
	return result, app.With(func(targets []evals.WorkflowTarget) error {
		index := slices.IndexFunc(targets, func(t evals.WorkflowTarget) bool {
			return t.Target().Name == r.target
		})
		if index < 0 {
			return fmt.Errorf("no evaluation target %q", r.target)
		}
		result = new(evals.RunWorkflow(ctx, targets[index], r.runs))
		return nil
	})
}
