package genkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/schema/schematypes"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/projections"
)

// fakeInstallation is an installed integration with only capabilities.
type fakeInstallation struct {
	intg         *ent.Integration
	capabilities []string
}

func newFakeInstallation(provider string, capabilities ...string) *fakeInstallation {
	intg := &ent.Integration{ID: uuid.New(), Provider: provider, Name: provider}
	return &fakeInstallation{intg: intg, capabilities: capabilities}
}

func (f *fakeInstallation) Integration() *ent.Integration             { return f.intg }
func (f *fakeInstallation) Config() rez.IntegrationInstallationConfig { return nil }
func (f *fakeInstallation) Capabilities() []string                    { return f.capabilities }

// fakeLogReader is an installation that reads logs labelled service. It reads a service by its first alias, else
// its display name, and records each read it is asked for.
type fakeLogReader struct {
	*fakeInstallation
	unavailable string
	readErr     error

	mu       sync.Mutex
	searches []integrations.LogSearchParams
	counts   []integrations.LogCountParams
}

// newFakeLogReader returns a log reader; a nonempty unavailable is why its logs cannot be read.
func newFakeLogReader(unavailable string) *fakeLogReader {
	return &fakeLogReader{fakeInstallation: newFakeInstallation("provider-a", "logs"), unavailable: unavailable}
}

func (f *fakeLogReader) CheckLogsAvailable() error {
	if f.unavailable != "" {
		return errors.New(f.unavailable)
	}
	return nil
}

func (f *fakeLogReader) serviceName(service integrations.LogService) integrations.LogServiceName {
	if len(service.Aliases) > 0 {
		return integrations.LogServiceName{Value: service.Aliases[0], FromAlias: true}
	}
	return integrations.LogServiceName{Value: service.DisplayName}
}

func (f *fakeLogReader) record(service integrations.LogServiceName, start, end time.Time) integrations.ReadRecord {
	query := fmt.Sprintf("{service=%q}", service.Value)
	return integrations.ReadRecord{
		DataSourceUID: "logs",
		Query:         query,
		Start:         start,
		End:           end,
		RanAt:         end,
		Limits:        []string{"reached the limit of 1 lines; more may exist"},
		ExploreURL:    "https://example.com/explore?query=" + query,
	}
}

func (f *fakeLogReader) SearchLogs(_ context.Context, params integrations.LogSearchParams) (*integrations.LogSearchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.searches = append(f.searches, params)
	if f.readErr != nil {
		return nil, f.readErr
	}
	service := f.serviceName(params.Service)
	line := integrations.LogLine{Time: params.End, Line: "upstream timeout", Labels: map[string]string{"service": service.Value}}
	result := &integrations.LogSearchResult{
		Service: service,
		Lines:   []integrations.LogLine{line},
		Read:    f.record(service, params.Start, params.End),
	}
	return result, nil
}

func (f *fakeLogReader) CountLogs(_ context.Context, params integrations.LogCountParams) (*integrations.LogCountResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.counts = append(f.counts, params)
	if f.readErr != nil {
		return nil, f.readErr
	}
	service := f.serviceName(params.Service)
	record := f.record(service, params.Start, params.End)
	record.Step = time.Minute
	return &integrations.LogCountResult{Service: service, Read: record}, nil
}

func (f *fakeLogReader) reads() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.searches) + len(f.counts)
}

// providerReadFixture is an open situation's investigation reading a checkout-api service through a log reader.
type providerReadFixture struct {
	h        *investigationHarness
	reader   *fakeLogReader
	service  *ent.KnowledgeEntity
	openedAt time.Time
}

func (s *AiRuntimeSuite) newProviderReadFixture() *providerReadFixture {
	h := s.newInvestigationHarness()
	openedAt := h.clock.Now().Add(-2 * time.Hour)
	h.invocation.providerReadWindow = &rezai.ProviderReadWindow{Start: openedAt.Add(-rezai.ProviderReadLeadIn)}
	service := &ent.KnowledgeEntity{
		ID:       uuid.New(),
		Category: kne.CategoryContainer,
		Kind:     "service",
		State:    schematypes.KnowledgeGraphSubjectState{DisplayName: "Checkout API"},
	}
	return &providerReadFixture{h: h, reader: newFakeLogReader(""), service: service, openedAt: openedAt}
}

// expectService makes the service readable, with the given aliases from the log reader's installation.
func (f *providerReadFixture) expectService(resourceRefs ...string) {
	f.h.knowledge.EXPECT().GetEntity(mock.Anything, f.service.ID, rez.KnowledgeEntityExpand{}).Return(f.service, nil)
	f.h.integrations.EXPECT().ListAllInstalled(mock.Anything).Return([]rez.InstalledIntegration{f.reader}, nil)
	aliases := &ent.ListResult[ent.KnowledgeSubjectAlias]{}
	for _, ref := range resourceRefs {
		aliases.Data = append(aliases.Data, &ent.KnowledgeSubjectAlias{ID: uuid.New(), ProviderResourceRef: ref})
	}
	f.h.knowledge.EXPECT().ListSubjectAliases(mock.Anything, mock.Anything).Return(aliases, nil)
}

// expectRecorded records reads as started with the given used count, and collects their outcomes.
func (f *providerReadFixture) expectRecorded(used int) (uuid.UUID, *[]rez.FinishAgentProviderReadParams) {
	readID := uuid.New()
	f.h.providerReads.EXPECT().StartProviderRead(mock.Anything, mock.Anything).
		Return(&ent.AgentProviderRead{ID: readID}, used, nil)
	var finished []rez.FinishAgentProviderReadParams
	f.h.providerReads.EXPECT().FinishProviderRead(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, params rez.FinishAgentProviderReadParams) error {
			finished = append(finished, params)
			return nil
		})
	return readID, &finished
}

func (f *providerReadFixture) search(start, end time.Time) rezai.SearchLogsArgs {
	return rezai.SearchLogsArgs{
		ServiceRef: f.service.ID.String(),
		Start:      start.Format(time.RFC3339),
		End:        end.Format(time.RFC3339),
	}
}

// runInvestigationToolResult runs the named tool through the error result middleware, as the agent's turn does,
// and decodes its result.
func (s *AiRuntimeSuite) runInvestigationToolResult(ctx context.Context, h *investigationHarness, name string, input any, output any) {
	hooks, hooksErr := (&toolErrorResultMiddleware{}).New(ctx)
	s.Require().NoError(hooksErr)
	for _, tool := range h.invocation.makeTools() {
		if tool.Name() != name {
			continue
		}
		params := &ai.ToolParams{Request: &ai.ToolRequest{Name: name, Input: input}, Tool: tool}
		resp, runErr := hooks.WrapTool(ctx, params, func(ctx context.Context, params *ai.ToolParams) (*ai.MultipartToolResponse, error) {
			return params.Tool.RunRawMultipart(ctx, params.Request.Input)
		})
		s.Require().NoError(runErr, "the turn continues")
		encoded, encodeErr := json.Marshal(resp.Output)
		s.Require().NoError(encodeErr)
		s.Require().NoError(json.Unmarshal(encoded, output))
		return
	}
	s.FailNow("tool not offered: " + name)
}

// requireErrorResult requires a result carrying only a nonempty error that contains the given text.
func (s *AiRuntimeSuite) requireErrorResult(result map[string]any, contains string) {
	s.T().Helper()
	errorText, isText := result["error"].(string)
	s.Require().True(isText && errorText != "" && len(result) == 1, "result %v carries only an error", result)
	s.Contains(errorText, contains)
}

func (s *AiRuntimeSuite) TestProviderReadsRefused() {
	ctx := s.T().Context()

	s.Run("before the window", func() {
		f := s.newProviderReadFixture()
		f.h.knowledge.EXPECT().GetEntity(mock.Anything, f.service.ID, rez.KnowledgeEntityExpand{}).Return(f.service, nil)
		start := f.openedAt.Add(-rezai.ProviderReadLeadIn - time.Minute)
		var result map[string]any
		s.runInvestigationToolResult(ctx, f.h, rezai.SearchLogsTool.Name(), f.search(start, start.Add(time.Hour)), &result)
		s.requireErrorResult(result, "window")
	})

	s.Run("after the situation closed", func() {
		f := s.newProviderReadFixture()
		closedAt := f.openedAt.Add(time.Hour)
		f.h.invocation.providerReadWindow.End = &closedAt
		f.h.knowledge.EXPECT().GetEntity(mock.Anything, f.service.ID, rez.KnowledgeEntityExpand{}).Return(f.service, nil)
		var result map[string]any
		s.runInvestigationToolResult(ctx, f.h, rezai.CountLogsTool.Name(), f.search(closedAt.Add(-time.Hour), closedAt.Add(time.Minute)), &result)
		s.requireErrorResult(result, "window")
	})

	s.Run("not a service", func() {
		f := s.newProviderReadFixture()
		f.service.Category = kne.CategorySignal
		f.service.Kind = "alert"
		f.h.knowledge.EXPECT().GetEntity(mock.Anything, f.service.ID, rez.KnowledgeEntityExpand{}).Return(f.service, nil)
		var result map[string]any
		s.runInvestigationToolResult(ctx, f.h, rezai.SearchLogsTool.Name(), f.search(f.openedAt, f.openedAt.Add(time.Hour)), &result)
		s.requireErrorResult(result, "not a service")
	})

	// An installation's own reason its logs are unavailable reaches the agent.
	unavailableReason := "reason-from-the-installation"
	for name, tc := range map[string]struct {
		installed []rez.InstalledIntegration
		contains  string
	}{
		"no logs integration installed": {installed: []rez.InstalledIntegration{newFakeInstallation("provider-b", "deployments")}},
		"logs unavailable":              {installed: []rez.InstalledIntegration{newFakeLogReader(unavailableReason)}, contains: unavailableReason},
	} {
		s.Run(name, func() {
			f := s.newProviderReadFixture()
			f.h.knowledge.EXPECT().GetEntity(mock.Anything, f.service.ID, rez.KnowledgeEntityExpand{}).Return(f.service, nil)
			f.h.integrations.EXPECT().ListAllInstalled(mock.Anything).Return(tc.installed, nil)
			var result map[string]any
			s.runInvestigationToolResult(ctx, f.h, rezai.SearchLogsTool.Name(), f.search(f.openedAt, f.openedAt.Add(time.Hour)), &result)
			s.requireErrorResult(result, tc.contains)
		})
	}
	// None of the refusals recorded a read: the provider read service mocks expect no calls.
}

func (s *AiRuntimeSuite) TestProviderReadsOfAnyService() {
	ctx := s.T().Context()
	f := s.newProviderReadFixture()
	// The service is not one of a situation's entities: the window is all the read is held to.
	f.expectService()
	readID, _ := f.expectRecorded(1)

	var result rezai.SearchLogsToolResult
	s.runInvestigationToolResult(ctx, f.h, rezai.SearchLogsTool.Name(), f.search(f.openedAt, f.openedAt.Add(time.Hour)), &result)

	s.Equal(readID.String(), result.ProviderReadRef)
	s.Equal(f.service.ID.String(), result.ServiceRef)
	s.Len(result.Lines, 1)
	s.Equal(rezai.MaxProviderReadsPerTurn-1, result.BudgetRemaining)
}

func (s *AiRuntimeSuite) TestProviderReadBudgetRefusalSkipsTheProvider() {
	ctx := s.T().Context()
	f := s.newProviderReadFixture()
	f.expectService()
	budgetErr := fmt.Errorf("%w: this turn has used its budget of %d provider reads", rezai.ErrToolRefused, rezai.MaxProviderReadsPerTurn)
	f.h.providerReads.EXPECT().StartProviderRead(mock.Anything, mock.Anything).Return(nil, 0, budgetErr)

	var result map[string]any
	s.runInvestigationToolResult(ctx, f.h, rezai.SearchLogsTool.Name(), f.search(f.openedAt, f.openedAt.Add(time.Hour)), &result)

	s.requireErrorResult(result, "budget")
	s.Zero(f.reader.reads(), "the provider is not called")
}

func (s *AiRuntimeSuite) TestProviderReadsAreRecorded() {
	ctx := s.T().Context()

	s.Run("succeeded", func() {
		f := s.newProviderReadFixture()
		f.expectService("checkout-api")
		readID, finished := f.expectRecorded(4)
		f.h.providerReads.EXPECT().StartProviderRead(mock.Anything, mock.Anything).Unset()
		f.h.providerReads.EXPECT().StartProviderRead(mock.Anything, mock.MatchedBy(func(params rez.StartAgentProviderReadParams) bool {
			return params.AgentTurnID == f.h.invocation.turnID &&
				params.ToolName == rezai.CountLogsTool.Name() &&
				params.ServiceEntityID == f.service.ID
		})).Return(&ent.AgentProviderRead{ID: readID}, 4, nil)

		input := rezai.CountLogsArgs{
			ServiceRef: f.service.ID.String(),
			Start:      f.openedAt.Format(time.RFC3339),
			End:        f.openedAt.Add(time.Hour).Format(time.RFC3339),
			Contains:   "timeout",
		}
		var result rezai.CountLogsToolResult
		s.runInvestigationToolResult(ctx, f.h, rezai.CountLogsTool.Name(), input, &result)

		s.Equal(readID.String(), result.ProviderReadRef)
		s.Equal(`{service="checkout-api"}`, result.Read.Query)
		s.Equal(f.openedAt, result.Read.Start)
		s.Equal("1m0s", result.Read.Step)
		s.NotEmpty(result.Read.ExploreURL)
		s.NotEmpty(result.Read.Limits)
		s.Equal(rezai.MaxProviderReadsPerTurn-4, result.BudgetRemaining)
		s.NotEmpty(result.Note, "a service with no lines says so")

		s.Require().Len(*finished, 1)
		outcome := (*finished)[0]
		s.Equal(readID, outcome.ID)
		s.Empty(outcome.Error)
		var storedRecord rezai.ProviderReadRecord
		s.Require().NoError(json.Unmarshal(outcome.Read, &storedRecord))
		s.Equal(result.Read, storedRecord)
		s.Equal(f.reader.Integration().Provider, storedRecord.Provider, "the record names the reading installation's provider")
		s.Equal(f.reader.Integration().Name, storedRecord.Integration, "and its integration")
		var storedResult rezai.CountLogsToolResult
		s.Require().NoError(json.Unmarshal(outcome.Result, &storedResult))
		s.Equal(result, storedResult, "the result is stored as returned to the agent")
	})

	s.Run("failed by the provider", func() {
		f := s.newProviderReadFixture()
		f.reader.readErr = errors.New("logs: 503 Service Unavailable")
		f.expectService()
		readID, finished := f.expectRecorded(1)

		var result map[string]any
		s.runInvestigationToolResult(ctx, f.h, rezai.SearchLogsTool.Name(), f.search(f.openedAt, f.openedAt.Add(time.Hour)), &result)

		s.requireErrorResult(result, "503 Service Unavailable")
		s.Require().Len(*finished, 1)
		s.Equal(readID, (*finished)[0].ID)
		s.Contains((*finished)[0].Error, "503 Service Unavailable")
		s.Empty((*finished)[0].Result)
	})

	s.Run("refused by a provider read limit", func() {
		f := s.newProviderReadFixture()
		f.expectService()
		search := f.search(f.openedAt, f.openedAt.Add(time.Hour))
		search.Limit = integrations.MaxLogSearchLimit + 1
		tooLong := f.search(f.openedAt.Add(-integrations.MaxTelemetryReadRange-time.Minute), f.openedAt)

		for _, input := range []rezai.SearchLogsArgs{search, tooLong} {
			var result map[string]any
			s.runInvestigationToolResult(ctx, f.h, rezai.SearchLogsTool.Name(), input, &result)
			s.requireErrorResult(result, "limit")
		}
		s.Zero(f.reader.reads())
		// Not recorded: the provider read service mock expects no calls.
	})
}

func (s *AiRuntimeSuite) TestProviderReadServiceIdentity() {
	ctx := s.T().Context()
	count := func(f *providerReadFixture) rezai.CountLogsToolResult {
		f.expectRecorded(1)
		input := rezai.CountLogsArgs{
			ServiceRef: f.service.ID.String(),
			Start:      f.openedAt.Format(time.RFC3339),
			End:        f.openedAt.Add(time.Hour).Format(time.RFC3339),
		}
		var result rezai.CountLogsToolResult
		s.runInvestigationToolResult(ctx, f.h, rezai.CountLogsTool.Name(), input, &result)
		return result
	}

	s.Run("the display name and the installation's aliases are passed", func() {
		f := s.newProviderReadFixture()
		f.expectService("checkout-api", "checkout")
		var aliasPredicates []predicate.KnowledgeSubjectAlias
		f.h.knowledge.EXPECT().ListSubjectAliases(mock.Anything, mock.Anything).Unset()
		f.h.knowledge.EXPECT().ListSubjectAliases(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, params rez.ListKnowledgeSubjectAliasesParams) (*ent.ListResult[ent.KnowledgeSubjectAlias], error) {
				aliasPredicates = params.Predicates
				return &ent.ListResult[ent.KnowledgeSubjectAlias]{Data: []*ent.KnowledgeSubjectAlias{
					{ID: uuid.New(), ProviderResourceRef: "checkout-api"},
					{ID: uuid.New(), ProviderResourceRef: "checkout"},
				}}, nil
			})

		result := count(f)

		s.Require().Len(f.reader.counts, 1)
		want := integrations.LogService{DisplayName: "Checkout API", Aliases: []string{"checkout-api", "checkout"}}
		s.Equal(want, f.reader.counts[0].Service)
		selector := sql.Select("*").From(sql.Table(ksa.Table))
		for _, aliasPredicate := range aliasPredicates {
			aliasPredicate(selector)
		}
		_, args := selector.Query()
		installation := f.reader.Integration()
		s.Equal([]any{f.service.ID, installation.Provider, installation.ID.String()}, args, "only the log installation's aliases")
		s.Equal("checkout-api", result.ServiceName, "read by the alias")
	})

	s.Run("without aliases, the display name is read by", func() {
		f := s.newProviderReadFixture()
		f.expectService()

		result := count(f)

		s.Equal("Checkout API", result.ServiceName)
		s.NotEmpty(result.Note)
		s.NotContains(result.Note, f.reader.Integration().Provider, "agent-facing text names no provider")
	})
}

func (s *AiRuntimeSuite) TestListChanges() {
	ctx := s.T().Context()
	h := s.newInvestigationHarness()
	deployedAt := time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC)
	deployment := &ent.KnowledgeEntity{ID: uuid.New(), Category: kne.CategoryEvent, Kind: "deployment"}
	change := &ent.KnowledgeEntity{
		ID:       uuid.New(),
		Category: kne.CategoryEvent,
		Kind:     "code_change",
		State:    schematypes.KnowledgeGraphSubjectState{DisplayName: "Shrink the payments pool"},
	}
	serviceID := uuid.New()
	params := rez.ListServiceChangesParams{
		ServiceEntityID: serviceID,
		Start:           deployedAt.Add(-time.Hour),
		End:             deployedAt.Add(time.Hour),
	}
	repositoryID := uuid.New()
	listed := &rez.ServiceChanges{Deployments: []rez.ServiceDeployment{{
		Deployment:                deployment,
		DeployedAt:                deployedAt,
		Environment:               "production",
		Status:                    "succeeded",
		Sha:                       "abc123",
		Version:                   "v42",
		URL:                       "https://ci.example.com/42",
		RepositoryEntityID:        &repositoryID,
		RepositoryName:            "acme/checkout",
		RepositorySeenByCodeForge: true,
		MergedChanges: []rez.LinkedMergedChange{{
			Change:   change,
			MergedAt: deployedAt.Add(-10 * time.Minute),
			Number:   7,
			Author:   "alice",
			Link:     rez.MergedChangeLinkMergeCommit,
		}},
	}}}
	h.changes.EXPECT().ListServiceChanges(mock.Anything, params).Return(listed, nil)
	deploymentAlias := &ent.KnowledgeSubjectAlias{ID: uuid.New(), EntityID: &deployment.ID}
	changeAlias := &ent.KnowledgeSubjectAlias{ID: uuid.New(), EntityID: &change.ID}
	aliases := &ent.ListResult[ent.KnowledgeSubjectAlias]{Data: []*ent.KnowledgeSubjectAlias{deploymentAlias, changeAlias}}
	h.knowledge.EXPECT().ListSubjectAliases(mock.Anything, mock.Anything).Return(aliases, nil)
	deploymentEvidence, changeEvidence := uuid.New(), uuid.New()
	summaries := &ent.ListResult[rez.KnowledgeEvidenceSummary]{Data: []*rez.KnowledgeEvidenceSummary{
		{SubjectAliasID: deploymentAlias.ID, Count: 2, LatestAt: deployedAt, LatestID: deploymentEvidence},
		{SubjectAliasID: changeAlias.ID, Count: 1, LatestAt: deployedAt, LatestID: changeEvidence},
	}}
	h.knowledge.EXPECT().SummarizeEvidence(mock.Anything, mock.Anything).Return(summaries, nil)

	input := rezai.ListChangesArgs{
		ServiceRef: serviceID.String(),
		Start:      params.Start.Format(time.RFC3339),
		End:        params.End.Format(time.RFC3339),
	}
	var result rezai.TextToolResult
	s.runInvestigationToolResult(ctx, h, rezai.ListChangesTool.Name(), input, &result)

	for _, ref := range []uuid.UUID{deployment.ID, deploymentEvidence, change.ID, changeEvidence} {
		s.Contains(result.Text, ref.String())
	}
	for _, field := range []string{"production", "succeeded", "abc123", "v42", "https://ci.example.com/42", "acme/checkout", "merge_commit", "7", "Shrink the payments pool", "alice"} {
		s.Contains(result.Text, field)
	}
	// Not a provider read: the provider read service mock expects no calls.

	s.Run("invalid service", func() {
		h.changes.EXPECT().ListServiceChanges(mock.Anything, mock.Anything).Unset()
		h.changes.EXPECT().ListServiceChanges(mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("%w: not a known service", errs.ErrInvalidInput))
		var invalid map[string]any
		s.runInvestigationToolResult(ctx, h, rezai.ListChangesTool.Name(), input, &invalid)
		s.requireErrorResult(invalid, "not a known service")
	})
}

func (s *AiRuntimeSuite) TestFindService() {
	ctx := s.T().Context()
	h := s.newInvestigationHarness()
	payments := &ent.KnowledgeEntity{
		ID:       uuid.New(),
		Category: kne.CategoryContainer,
		Kind:     "service",
		State:    schematypes.KnowledgeGraphSubjectState{DisplayName: "payments-api"},
	}
	byName := func(name string) rez.ListKnowledgeEntitiesParams {
		return rez.ListKnowledgeEntitiesParams{
			ListParams:       ent.ListParams{PageSize: 1},
			Categories:       []kne.Category{kne.CategoryContainer},
			LinkingAttribute: projections.LinkingAttributeServiceName,
			LinkingValue:     name,
		}
	}
	found := &ent.ListResult[ent.KnowledgeEntity]{Data: []*ent.KnowledgeEntity{payments}, Total: 1}
	h.knowledge.EXPECT().ListEntities(mock.Anything, byName("payments-api")).Return(found, nil)
	h.knowledge.EXPECT().ListEntities(mock.Anything, byName("orders-api")).Return(&ent.ListResult[ent.KnowledgeEntity]{}, nil)

	var result rezai.TextToolResult
	s.runInvestigationToolResult(ctx, h, rezai.FindServiceTool.Name(), rezai.FindServiceArgs{Name: "Payments_API"}, &result)
	s.Contains(result.Text, payments.ID.String())

	var unknown rezai.TextToolResult
	s.runInvestigationToolResult(ctx, h, rezai.FindServiceTool.Name(), rezai.FindServiceArgs{Name: "orders-api"}, &unknown)
	s.NotEmpty(unknown.Text)
	// Only containers holding the name's service.name linking attribute are listed: see the query service's
	// TestListEntitiesByLinkingAttribute.
}
