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
	"github.com/samber/do/v2"
)

// pkgEvals provides the evaluation integration, and the targets and services the evals commands run. It is part
// of every run: the integration and its event processor are always registered, but only agentEvaluationProfile
// makes the integration available.
var pkgEvals = do.Package(
	do.Lazy(func(i do.Injector) (*evals.Integration, error) {
		return evals.NewIntegration(do.MustInvoke[rez.Clock](i), false), nil
	}),

	do.Lazy(func(i do.Injector) ([]evals.WorkflowTarget, error) {
		workflows := evals.Workflows{
			JudgeSituationCandidate: do.MustInvoke[rezai.JudgeSituationCandidateWorkflow](i),
		}
		return workflows.Targets()
	}),

	do.Lazy(func(i do.Injector) (*evals.AgentEvalService, error) {
		return evals.NewAgentEvalService(
			do.MustInvoke[rez.Database](i),
			do.MustInvoke[rez.Clock](i),
			do.MustInvoke[rez.AuthSessionService](i),
			do.MustInvoke[rez.IntegrationService](i),
			do.MustInvoke[rez.ProviderEventIngester](i),
			do.MustInvoke[rez.EventProjectionService](i),
			do.MustInvoke[rez.SituationService](i),
			do.MustInvoke[rez.InvestigationService](i),
		), nil
	}),
)

// agentEvaluationProfile prepares the application for agent evaluation scenarios, on a test database: the
// evaluation integration is available, setup events skip detection, and neither periodic jobs nor the model
// judge run, since the scenario decides what the agent faces.
var agentEvaluationProfile = do.Package(
	testDatabaseProfile,
	noPeriodicJobsProfile,
	override(func(i do.Injector) (*evals.Integration, error) {
		return evals.NewIntegration(do.MustInvoke[rez.Clock](i), true), nil
	}),
	override(func(i do.Injector) (rez.SituationSignalService, error) {
		return evals.NewSetupSignals(do.MustInvoke[*db.SituationSignalService](i)), nil
	}),
	override(func(i do.Injector) (rez.SituationsConfig, error) {
		situations := do.MustInvoke[rez.Config](i).Situations
		situations.ModelJudge = false
		return situations, nil
	}),
)

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
		profiles = append(profiles, agentEvaluationProfile)
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
