package integrations

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

const (
	// CapabilityDeployments is what an installation that reports deployments, such as a webhook with the
	// deployment preset, says it provides.
	CapabilityDeployments = "deployments"
	// CapabilityCodeChanges is what an installation that reports merged code changes, such as a code forge,
	// says it provides.
	CapabilityCodeChanges = "code_changes"
)

// Sources says which sources the organization's installations can read, and why not.
type Sources struct {
	Logs    rezai.SourceAvailability
	Changes rezai.SourceAvailability
	// CodeChanges is whether merged code changes are reported, so deployments can be linked to them.
	CodeChanges rezai.SourceAvailability
}

// CheckSources says which sources can be read, and why not, without calling a provider. An installation that
// cannot be loaded makes the sources unavailable rather than failing.
func CheckSources(ctx context.Context, getter rez.InstalledIntegrationGetter) Sources {
	installed, listErr := getter.ListAllInstalled(ctx)
	if listErr != nil {
		slog.ErrorContext(ctx, "list installed integrations for sources", "error", listErr)
		unavailable := rezai.SourceAvailability{Reason: "installed integrations could not be loaded"}
		return Sources{Logs: unavailable, Changes: unavailable, CodeChanges: unavailable}
	}
	sources := Sources{
		Logs:        rezai.SourceAvailability{Available: true},
		Changes:     rezai.SourceAvailability{Reason: "no deployments integration installed"},
		CodeChanges: rezai.SourceAvailability{Reason: "no code changes integration installed"},
	}
	if _, logsErr := LookupLogSource(ctx, getter); logsErr != nil {
		var unavailable *UnavailableError
		if !errors.As(logsErr, &unavailable) {
			slog.ErrorContext(ctx, "look up log source for sources", "error", logsErr)
			unavailable = &UnavailableError{Reason: "the logs integration could not be loaded"}
		}
		sources.Logs = rezai.SourceAvailability{Reason: unavailable.Reason}
	}
	for _, ii := range installed {
		if slices.Contains(ii.Capabilities(), CapabilityDeployments) {
			sources.Changes = rezai.SourceAvailability{Available: true}
		}
		if slices.Contains(ii.Capabilities(), CapabilityCodeChanges) {
			sources.CodeChanges = rezai.SourceAvailability{Available: true}
		}
	}
	return sources
}
