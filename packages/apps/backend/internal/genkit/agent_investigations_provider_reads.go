package genkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/projections"
)

// deploymentsCapability is what an installation that reports deployments, such as a webhook with the
// deployment preset, says it provides.
const deploymentsCapability = "deployments"

// introduceSources says which sources the agent can read, and why not, without calling a provider. An
// installation that cannot be loaded makes the sources unavailable rather than failing the turn.
func (i *investigationInvocation) introduceSources(ctx context.Context) *rezai.InvestigationSources {
	installed, listErr := i.integrations.ListAllInstalled(ctx)
	if listErr != nil {
		slog.ErrorContext(ctx, "list installed integrations for investigation sources", "error", listErr)
		unavailable := rezai.SourceAvailability{Reason: "installed integrations could not be loaded"}
		return &rezai.InvestigationSources{Logs: unavailable, Changes: unavailable}
	}
	sources := &rezai.InvestigationSources{
		Logs:    rezai.SourceAvailability{Available: true},
		Changes: rezai.SourceAvailability{Reason: "no deployments integration installed"},
	}
	if _, logsErr := integrations.LookupLogSource(ctx, i.integrations); logsErr != nil {
		var unavailable *integrations.UnavailableError
		if !errors.As(logsErr, &unavailable) {
			slog.ErrorContext(ctx, "look up log source for investigation sources", "error", logsErr)
			unavailable = &integrations.UnavailableError{Reason: "the logs integration could not be loaded"}
		}
		sources.Logs = rezai.SourceAvailability{Reason: unavailable.Reason}
	}
	for _, ii := range installed {
		if slices.Contains(ii.Capabilities(), deploymentsCapability) {
			sources.Changes = rezai.SourceAvailability{Available: true}
		}
	}
	return sources
}

// findService returns the service holding the name's service.name linking attribute. A linking attribute value
// identifies one entity, so there is one service or none.
func (i *investigationInvocation) findService(ctx context.Context, input rezai.FindServiceArgs) (*rezai.TextToolResult, error) {
	name := projections.NormalizeServiceName(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name %q has no name once normalized", errs.ErrInvalidInput, input.Name)
	}
	params := rez.ListKnowledgeEntitiesParams{
		ListParams:       ent.ListParams{PageSize: 1},
		Categories:       []kne.Category{kne.CategoryContainer},
		LinkingAttribute: projections.LinkingAttributeServiceName,
		LinkingValue:     name,
	}
	services, listErr := i.knowledge.ListEntities(ctx, params)
	if listErr != nil {
		return nil, fmt.Errorf("find service: %w", listErr)
	}
	if len(services.Data) == 0 {
		return &rezai.TextToolResult{Text: fmt.Sprintf("No service is named %q.", name)}, nil
	}
	return &rezai.TextToolResult{Text: i.renderEntityLine(services.Data[0])}, nil
}

func (i *investigationInvocation) listChanges(ctx context.Context, input rezai.ListChangesArgs) (*rezai.TextToolResult, error) {
	serviceID, resolveErr := i.resolveRef(input.ServiceRef)
	if resolveErr != nil {
		return nil, fmt.Errorf("service_ref: %w", resolveErr)
	}
	start, end, rangeErr := i.parseRange(input.Start, input.End)
	if rangeErr != nil {
		return nil, rangeErr
	}
	params := rez.ListServiceChangesParams{
		ServiceEntityID: serviceID,
		Start:           start,
		End:             end,
		Environment:     input.Environment,
	}
	changes, changesErr := i.changes.ListServiceChanges(ctx, params)
	if changesErr != nil {
		return nil, fmt.Errorf("list service changes: %w", changesErr)
	}

	var entities []*ent.KnowledgeEntity
	for _, deployment := range changes.Deployments {
		entities = append(entities, deployment.Deployment)
		for _, change := range deployment.MergedChanges {
			entities = append(entities, change.Change)
		}
	}
	latestEvidence, evidenceErr := i.latestEvidenceByEntity(ctx, entities)
	if evidenceErr != nil {
		return nil, fmt.Errorf("summarize deployment and change evidence: %w", evidenceErr)
	}

	var lines []string
	for _, deployment := range changes.Deployments {
		lines = append(lines, i.renderDeploymentLine(deployment, latestEvidence))
		for _, change := range deployment.MergedChanges {
			lines = append(lines, "  "+i.renderMergedChangeLine(change, latestEvidence))
		}
		if deployment.MergedChangesCut {
			lines = append(lines, "  merged changes cut: more were linked than are listed")
		}
	}
	if len(lines) == 0 {
		lines = append(lines, "No deployments in the window.")
	}
	for _, limit := range changes.Limits {
		lines = append(lines, "limit: "+limit)
	}
	return &rezai.TextToolResult{Text: limitedAnalysisResultText(strings.Join(lines, "\n"), "")}, nil
}

func (i *investigationInvocation) renderDeploymentLine(deployment rez.ServiceDeployment, latestEvidence map[uuid.UUID]uuid.UUID) string {
	fields := []string{
		"deployment " + i.formatRef(deployment.Deployment.ID),
		"deployed_at=" + deployment.DeployedAt.UTC().Format(time.RFC3339Nano),
		"environment=" + deployment.Environment,
		"status=" + deployment.Status,
		"commit=" + deployment.Sha,
		"version=" + deployment.Version,
		"url=" + deployment.URL,
	}
	switch {
	case deployment.RepositoryEntityID == nil:
		fields = append(fields, "repository=none (no merged changes can be linked)")
	case !deployment.RepositorySeenByCodeForge:
		fields = append(fields, fmt.Sprintf("repository=%q (not seen by a code forge, so no merged changes can be linked)", deployment.RepositoryName))
	default:
		fields = append(fields, fmt.Sprintf("repository=%q", deployment.RepositoryName))
	}
	if deployment.PreviousDeployedAt != nil {
		fields = append(fields, "previous_deployed_at="+deployment.PreviousDeployedAt.UTC().Format(time.RFC3339Nano))
	}
	fields = append(fields, i.renderLatestEvidence(deployment.Deployment.ID, latestEvidence))
	return strings.Join(fields, " ")
}

func (i *investigationInvocation) renderMergedChangeLine(change rez.LinkedMergedChange, latestEvidence map[uuid.UUID]uuid.UUID) string {
	fields := []string{
		"merged_change " + i.formatRef(change.Change.ID),
		"number=" + strconv.Itoa(change.Number),
		fmt.Sprintf("title=%q", limitedAnalysisDescription(change.Change.State.DisplayName)),
		"author=" + change.Author,
		"merged_at=" + change.MergedAt.UTC().Format(time.RFC3339Nano),
		"link=" + string(change.Link),
		i.renderLatestEvidence(change.Change.ID, latestEvidence),
	}
	return strings.Join(fields, " ")
}

func (i *investigationInvocation) renderLatestEvidence(entityID uuid.UUID, latestEvidence map[uuid.UUID]uuid.UUID) string {
	evidenceID, found := latestEvidence[entityID]
	if !found {
		return "latest_evidence=none"
	}
	return "latest_evidence=" + i.formatRef(evidenceID)
}

// latestEvidenceByEntity returns the latest evidence on each entity's aliases, by entity ID.
func (i *investigationInvocation) latestEvidenceByEntity(ctx context.Context, entities []*ent.KnowledgeEntity) (map[uuid.UUID]uuid.UUID, error) {
	latest := make(map[uuid.UUID]uuid.UUID, len(entities))
	if len(entities) == 0 {
		return latest, nil
	}
	entityIDs := make([]uuid.UUID, len(entities))
	for index, entity := range entities {
		entityIDs[index] = entity.ID
	}
	// An event entity has an alias per source that reported it, so this is a few per entity.
	aliasesParams := rez.ListKnowledgeSubjectAliasesParams{
		ListParams: ent.ListParams{PageSize: 10 * len(entityIDs)},
		Predicates: []predicate.KnowledgeSubjectAlias{ksa.EntityIDIn(entityIDs...)},
	}
	aliases, aliasesErr := i.knowledge.ListSubjectAliases(ctx, aliasesParams)
	if aliasesErr != nil {
		return nil, fmt.Errorf("list entity aliases: %w", aliasesErr)
	}
	summaries, summarizeErr := i.summarizeAliasEvidence(ctx, aliases.Data)
	if summarizeErr != nil {
		return nil, summarizeErr
	}
	latestAt := make(map[uuid.UUID]time.Time, len(entities))
	for _, alias := range aliases.Data {
		summary, found := summaries[alias.ID]
		if !found || alias.EntityID == nil {
			continue
		}
		if previous, seen := latestAt[*alias.EntityID]; !seen || summary.LatestAt.After(previous) {
			latestAt[*alias.EntityID] = summary.LatestAt
			latest[*alias.EntityID] = summary.LatestID
		}
	}
	return latest, nil
}

// logRead is a validated log read: the service, the range and the service's identity for the log source.
type logRead struct {
	service    *ent.KnowledgeEntity
	source     *integrations.LogSource
	identity   integrations.LogService
	start, end time.Time
}

func (i *investigationInvocation) searchLogs(ctx context.Context, input rezai.SearchLogsArgs) (*rezai.SearchLogsToolResult, error) {
	target, prepareErr := i.prepareLogRead(ctx, input.ServiceRef, input.Start, input.End)
	if prepareErr != nil {
		return nil, prepareErr
	}
	params := integrations.LogSearchParams{
		Service:  target.identity,
		Start:    target.start,
		End:      target.end,
		Contains: input.Contains,
		Limit:    input.Limit,
	}
	if validateErr := params.Validate(); validateErr != nil {
		return nil, validateErr
	}

	run, startErr := i.startProviderRead(ctx, rezai.SearchLogsTool.Name(), target.service.ID, input)
	if startErr != nil {
		return nil, startErr
	}
	searched, searchErr := target.source.Querier.SearchLogs(ctx, params)
	if searchErr != nil {
		return nil, i.failProviderRead(ctx, run, searchErr)
	}
	result := &rezai.SearchLogsToolResult{
		ProviderReadRef: i.formatRef(run.id),
		ServiceRef:      i.formatRef(target.service.ID),
		ServiceName:     searched.Service.Value,
		Lines:           make([]rezai.ProviderReadLogLine, len(searched.Lines)),
		Read:            i.providerReadRecord(target.source, searched.Read),
		BudgetRemaining: run.budgetRemaining,
	}
	for index, line := range searched.Lines {
		result.Lines[index] = rezai.ProviderReadLogLine{Time: line.Time, Line: line.Line, Labels: line.Labels}
	}
	if len(result.Lines) == 0 {
		result.Note = noLogsNote("lines", searched.Service)
	}
	if finishErr := i.finishProviderRead(ctx, run, result.Read, result); finishErr != nil {
		return nil, finishErr
	}
	return result, nil
}

func (i *investigationInvocation) countLogs(ctx context.Context, input rezai.CountLogsArgs) (*rezai.CountLogsToolResult, error) {
	target, prepareErr := i.prepareLogRead(ctx, input.ServiceRef, input.Start, input.End)
	if prepareErr != nil {
		return nil, prepareErr
	}
	params := integrations.LogCountParams{
		Service:  target.identity,
		Start:    target.start,
		End:      target.end,
		Contains: input.Contains,
	}
	if validateErr := params.Validate(); validateErr != nil {
		return nil, validateErr
	}

	run, startErr := i.startProviderRead(ctx, rezai.CountLogsTool.Name(), target.service.ID, input)
	if startErr != nil {
		return nil, startErr
	}
	counted, countErr := target.source.Querier.CountLogs(ctx, params)
	if countErr != nil {
		return nil, i.failProviderRead(ctx, run, countErr)
	}
	result := &rezai.CountLogsToolResult{
		ProviderReadRef: i.formatRef(run.id),
		ServiceRef:      i.formatRef(target.service.ID),
		ServiceName:     counted.Service.Value,
		Points:          make([]rezai.ProviderReadPoint, len(counted.Points)),
		Read:            i.providerReadRecord(target.source, counted.Read),
		BudgetRemaining: run.budgetRemaining,
	}
	hasLines := false
	for index, point := range counted.Points {
		result.Points[index] = rezai.ProviderReadPoint{Time: point.Time, Value: point.Value}
		hasLines = hasLines || point.Value > 0
	}
	if !hasLines {
		result.Note = noLogsNote("lines counted", counted.Service)
	}
	if finishErr := i.finishProviderRead(ctx, run, result.Read, result); finishErr != nil {
		return nil, finishErr
	}
	return result, nil
}

// noLogsNote says that the service had no lines, and why that may be when it was read by its display name.
func noLogsNote(what string, service integrations.LogServiceName) string {
	if service.FromAlias {
		return fmt.Sprintf("No %s for %q in this range.", what, service.Value)
	}
	return fmt.Sprintf("No %s for %q in this range. The logs integration has not reported this service's name, so its display name was used; its logs may carry another name.",
		what, service.Value)
}

// prepareLogRead validates a log read before it is recorded: the service, the window and the log source. A
// refusal here does not count against the budget.
func (i *investigationInvocation) prepareLogRead(ctx context.Context, serviceRef, rawStart, rawEnd string) (*logRead, error) {
	serviceID, resolveErr := i.resolveRef(serviceRef)
	if resolveErr != nil {
		return nil, fmt.Errorf("service_ref: %w", resolveErr)
	}
	start, end, rangeErr := i.parseRange(rawStart, rawEnd)
	if rangeErr != nil {
		return nil, rangeErr
	}
	service, serviceErr := i.knowledge.GetEntity(ctx, serviceID, rez.KnowledgeEntityExpand{})
	if serviceErr != nil {
		return nil, fmt.Errorf("read service: %w", serviceErr)
	}
	if service.Category != kne.CategoryContainer {
		return nil, fmt.Errorf("%w: service_ref %s is a %s/%s, not a service", errs.ErrInvalidInput, serviceRef, service.Category, service.Kind)
	}
	if windowErr := i.checkProviderReadWindow(start, end); windowErr != nil {
		return nil, windowErr
	}
	source, sourceErr := integrations.LookupLogSource(ctx, i.integrations)
	if sourceErr != nil {
		var unavailable *integrations.UnavailableError
		if errors.As(sourceErr, &unavailable) {
			return nil, fmt.Errorf("%w: logs cannot be read: %s", rezai.ErrToolRefused, unavailable.Reason)
		}
		return nil, fmt.Errorf("look up log source: %w", sourceErr)
	}

	aliases, aliasesErr := i.logSourceAliases(ctx, service.ID, source)
	if aliasesErr != nil {
		return nil, aliasesErr
	}
	identity := integrations.LogService{DisplayName: service.State.DisplayName, Aliases: aliases}
	return &logRead{service: service, source: source, identity: identity, start: start, end: end}, nil
}

// logSourceAliases returns the resource refs of the service's aliases from the log source's installation, for
// the integration to choose the value its logs are read by.
func (i *investigationInvocation) logSourceAliases(ctx context.Context, serviceID uuid.UUID, source *integrations.LogSource) ([]string, error) {
	params := rez.ListKnowledgeSubjectAliasesParams{
		ListParams: ent.ListParams{PageSize: 50},
		Predicates: []predicate.KnowledgeSubjectAlias{
			ksa.EntityID(serviceID),
			ksa.Provider(source.Installation.Provider),
			ksa.ProviderNamespace(source.Installation.ID.String()),
		},
	}
	aliases, listErr := i.knowledge.ListSubjectAliases(ctx, params)
	if listErr != nil {
		return nil, fmt.Errorf("list service aliases from the log source: %w", listErr)
	}
	refs := make([]string, len(aliases.Data))
	for index, alias := range aliases.Data {
		refs[index] = alias.ProviderResourceRef
	}
	return refs, nil
}

// checkProviderReadWindow refuses a range outside the window: from ProviderReadLeadIn before the situation
// opened until it closed, or until now while it is open.
func (i *investigationInvocation) checkProviderReadWindow(start, end time.Time) error {
	window := i.providerReadWindow
	if window == nil {
		return fmt.Errorf("%w: this investigation has no situation, so it has no provider read window", rezai.ErrToolRefused)
	}
	windowEnd := i.clock.Now()
	if window.End != nil {
		windowEnd = *window.End
	}
	if start.Before(window.Start) {
		return fmt.Errorf("%w: start %s is before the provider read window, which starts at %s",
			rezai.ErrToolRefused, start.UTC().Format(time.RFC3339), window.Start.UTC().Format(time.RFC3339))
	}
	if end.After(windowEnd) {
		return fmt.Errorf("%w: end %s is after the provider read window, which ends at %s",
			rezai.ErrToolRefused, end.UTC().Format(time.RFC3339), windowEnd.UTC().Format(time.RFC3339))
	}
	return nil
}

// providerReadRun is a provider read recorded as running.
type providerReadRun struct {
	id              uuid.UUID
	budgetRemaining int
}

func (i *investigationInvocation) startProviderRead(ctx context.Context, toolName string, serviceID uuid.UUID, input any) (*providerReadRun, error) {
	encodedInput, encodeErr := json.Marshal(input)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode provider read input: %w", encodeErr)
	}
	params := rez.StartAgentProviderReadParams{
		AgentTurnID:     i.turnID,
		ToolName:        toolName,
		ServiceEntityID: serviceID,
		Input:           encodedInput,
	}
	read, used, startErr := i.providerReads.StartProviderRead(ctx, params)
	if startErr != nil {
		return nil, fmt.Errorf("start provider read: %w", startErr)
	}
	return &providerReadRun{id: read.ID, budgetRemaining: rezai.MaxProviderReadsPerTurn - used}, nil
}

func (i *investigationInvocation) finishProviderRead(ctx context.Context, run *providerReadRun, record rezai.ProviderReadRecord, result any) error {
	encodedRecord, recordErr := json.Marshal(record)
	if recordErr != nil {
		return fmt.Errorf("encode provider read record: %w", recordErr)
	}
	encodedResult, resultErr := json.Marshal(result)
	if resultErr != nil {
		return fmt.Errorf("encode provider read result: %w", resultErr)
	}
	params := rez.FinishAgentProviderReadParams{ID: run.id, Read: encodedRecord, Result: encodedResult}
	if finishErr := i.providerReads.FinishProviderRead(ctx, params); finishErr != nil {
		return fmt.Errorf("record provider read result: %w", finishErr)
	}
	return nil
}

// failProviderRead records a provider's failure and returns it to the agent as a refusal. A read cut short by
// the turn ending stays running and fails the turn.
func (i *investigationInvocation) failProviderRead(ctx context.Context, run *providerReadRun, readErr error) error {
	if ctx.Err() != nil {
		return fmt.Errorf("provider read: %w", readErr)
	}
	params := rez.FinishAgentProviderReadParams{ID: run.id, Error: readErr.Error()}
	if finishErr := i.providerReads.FinishProviderRead(ctx, params); finishErr != nil {
		return fmt.Errorf("record provider read failure: %w", finishErr)
	}
	return fmt.Errorf("%w: the provider read failed: %w", rezai.ErrToolRefused, readErr)
}

func (i *investigationInvocation) providerReadRecord(source *integrations.LogSource, read integrations.ReadRecord) rezai.ProviderReadRecord {
	record := rezai.ProviderReadRecord{
		Provider:    source.Installation.Provider,
		Integration: source.Installation.Name,
		Query:       read.Query,
		Start:       read.Start,
		End:         read.End,
		RanAt:       read.RanAt,
		Limits:      read.Limits,
		ExploreURL:  read.ExploreURL,
	}
	if record.Limits == nil {
		record.Limits = []string{}
	}
	if read.Step > 0 {
		record.Step = read.Step.String()
	}
	return record
}

// parseRange parses a required RFC 3339 start and end.
func (i *investigationInvocation) parseRange(rawStart, rawEnd string) (time.Time, time.Time, error) {
	if rawStart == "" || rawEnd == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: start and end are required", errs.ErrInvalidInput)
	}
	start, startErr := i.parseOptionalTime("start", rawStart)
	if startErr != nil {
		return time.Time{}, time.Time{}, startErr
	}
	end, endErr := i.parseOptionalTime("end", rawEnd)
	if endErr != nil {
		return time.Time{}, time.Time{}, endErr
	}
	return start, end, nil
}
