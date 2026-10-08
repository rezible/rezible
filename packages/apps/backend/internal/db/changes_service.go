package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	changesMaxWindow               = 7 * 24 * time.Hour
	changesMaxDeployments          = 20
	changesMaxMergedChanges        = 20
	changesLookbackWithoutPrevious = 7 * 24 * time.Hour
	changesServiceEntityKind       = "service"
	changesRepositoryEntityKind    = "repository"
	changesDeploymentEventKind     = "deployment"
	changesCodeChangeEventKind     = "code_change"
	// changesDeploymentReportProvider is the webhook integration's provider name. A repository with an alias
	// from any other provider has been seen by a code forge.
	changesDeploymentReportProvider = "webhook"
)

type ChangesService struct {
	db    rez.Database
	graph rez.KnowledgeGraphQueryService
}

func NewChangesService(db rez.Database, graph rez.KnowledgeGraphQueryService) (*ChangesService, error) {
	return &ChangesService{db: db, graph: graph}, nil
}

func (s *ChangesService) ListServiceChanges(ctx context.Context, params rez.ListServiceChangesParams) (*rez.ServiceChanges, error) {
	if !params.End.After(params.Start) {
		return nil, fmt.Errorf("%w: end must be after start", rez.ErrInvalidInput)
	}
	if params.End.Sub(params.Start) > changesMaxWindow {
		return nil, fmt.Errorf("%w: the window is longer than the 7 day limit", rez.ErrInvalidInput)
	}
	environment := ""
	if params.Environment != "" {
		environment = projections.NormalizeServiceName(params.Environment)
		if environment == "" {
			return nil, fmt.Errorf("%w: environment %q has no name once normalized", rez.ErrInvalidInput, params.Environment)
		}
	}

	queryService := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(
			kne.ID(params.ServiceEntityID),
			kne.CategoryEQ(kne.CategoryContainer),
			kne.Kind(changesServiceEntityKind),
		)
	isService, serviceErr := queryService.Exist(ctx)
	if serviceErr != nil {
		return nil, fmt.Errorf("query service entity: %w", serviceErr)
	} else if !isService {
		return nil, fmt.Errorf("%w: %s is not a known service", rez.ErrInvalidInput, params.ServiceEntityID)
	}

	listDeployments := rez.ListRelatedEventsParams{
		EntityID:      params.ServiceEntityID,
		Predicate:     knr.PredicateImpacts,
		Kind:          changesDeploymentEventKind,
		From:          params.Start,
		FromInclusive: true,
		To:            params.End,
		ToInclusive:   true,
		Limit:         changesMaxDeployments,
	}
	if environment != "" {
		listDeployments.PropertyEquals = map[string]string{
			projections.DeploymentPropertyEnvironment: environment,
		}
	}
	deployments, deploymentsErr := s.graph.ListRelatedEvents(ctx, listDeployments)
	if deploymentsErr != nil {
		return nil, fmt.Errorf("list deployments: %w", deploymentsErr)
	}

	result := &rez.ServiceChanges{}
	if deployments.Cut {
		limit := fmt.Sprintf("more than %d deployments in the window; listing the %d most recent", changesMaxDeployments, changesMaxDeployments)
		result.Limits = append(result.Limits, limit)
	}
	for _, d := range deployments.Events {
		listed, listErr := s.listDeploymentChanges(ctx, params.ServiceEntityID, d)
		if listErr != nil {
			return nil, fmt.Errorf("deployment %s: %w", d.ID, listErr)
		}
		result.Deployments = append(result.Deployments, *listed)
	}
	return result, nil
}

func (s *ChangesService) listDeploymentChanges(ctx context.Context, serviceEntityID uuid.UUID, d *ent.KnowledgeEntity) (*rez.ServiceDeployment, error) {
	properties := changesEventProperties(d.State.Properties)
	listed := &rez.ServiceDeployment{
		Deployment:  d,
		DeployedAt:  *d.StateEffectiveAt,
		Environment: properties.get(projections.DeploymentPropertyEnvironment),
		Status:      properties.get(projections.DeploymentPropertyStatus),
		Sha:         properties.get(projections.DeploymentPropertySha),
		Version:     properties.get(projections.DeploymentPropertyVersion),
		URL:         properties.get(projections.DeploymentPropertyURL),
	}
	repositoryFullName := properties.get(projections.DeploymentPropertyRepository)

	listPrevious := rez.ListRelatedEventsParams{
		EntityID:  serviceEntityID,
		Predicate: knr.PredicateImpacts,
		Kind:      changesDeploymentEventKind,
		To:        listed.DeployedAt,
		PropertyEquals: map[string]string{
			projections.DeploymentPropertyStatus:      projections.DeploymentStatusSucceeded,
			projections.DeploymentPropertyEnvironment: listed.Environment,
		},
		Limit: 1,
	}
	if repositoryFullName != "" {
		listPrevious.PropertyEquals[projections.DeploymentPropertyRepository] = repositoryFullName
	} else {
		listPrevious.PropertyAbsent = []string{projections.DeploymentPropertyRepository}
	}
	previous, previousErr := s.graph.ListRelatedEvents(ctx, listPrevious)
	if previousErr != nil {
		return nil, fmt.Errorf("list previous deployment: %w", previousErr)
	}
	if len(previous.Events) > 0 {
		listed.PreviousDeployedAt = previous.Events[0].StateEffectiveAt
	}

	if repositoryFullName == "" {
		return listed, nil
	}
	repository, repositoryErr := s.queryDeploymentRepository(ctx, d.ID, repositoryFullName)
	if repositoryErr != nil {
		return nil, fmt.Errorf("query repository: %w", repositoryErr)
	}
	listed.RepositoryEntityID = &repository.ID
	listed.RepositoryName = repository.State.DisplayName
	listed.RepositorySeenByCodeForge = len(repository.Edges.Aliases) > 0

	if !listed.RepositorySeenByCodeForge {
		return listed, nil
	}
	changes, cut, changesErr := s.linkMergedChanges(ctx, listed)
	if changesErr != nil {
		return nil, fmt.Errorf("link merged changes: %w", changesErr)
	}
	listed.MergedChanges = changes
	listed.MergedChangesCut = cut
	return listed, nil
}

// queryDeploymentRepository returns the repository the deployment touches under the name its state reports,
// with the repository's aliases from code forges.
func (s *ChangesService) queryDeploymentRepository(ctx context.Context, deploymentID uuid.UUID, fullName string) (*ent.KnowledgeEntity, error) {
	queryRepository := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(
			kne.CategoryEQ(kne.CategoryCode),
			kne.Kind(changesRepositoryEntityKind),
			kne.HasTargetRelationshipsWith(
				knr.PredicateEQ(knr.PredicateTouches),
				knr.SourceEntityID(deploymentID),
			),
			kne.HasLinkingAttributesWith(
				kela.Attribute(projections.LinkingAttributeRepositoryFullName),
				kela.Value(fullName),
			),
		).
		WithAliases(func(q *ent.KnowledgeSubjectAliasQuery) {
			q.Where(ksa.ProviderNEQ(changesDeploymentReportProvider))
		})
	return queryRepository.Only(ctx)
}

// linkMergedChanges lists the merged changes of the deployment's repository linked to it: merge commit
// matches first, then merges into the default branch since the previous deployment, or within the lookback
// when there is none. Merge commit matches are always kept; the rest fill the remaining slots. It reports
// whether the limit cut the list.
func (s *ChangesService) linkMergedChanges(ctx context.Context, listed *rez.ServiceDeployment) ([]rez.LinkedMergedChange, bool, error) {
	var linked []rez.LinkedMergedChange
	matched := mapset.NewSet[uuid.UUID]()
	cut := false

	if listed.Sha != "" {
		listMatches := rez.ListRelatedEventsParams{
			EntityID:  *listed.RepositoryEntityID,
			Predicate: knr.PredicateTouches,
			Kind:      changesCodeChangeEventKind,
			PropertyEquals: map[string]string{
				projections.CodeChangePropertyMergeCommitSha: listed.Sha,
			},
			Limit: changesMaxMergedChanges,
		}
		matches, matchesErr := s.graph.ListRelatedEvents(ctx, listMatches)
		if matchesErr != nil {
			return nil, false, fmt.Errorf("list merge commit matches: %w", matchesErr)
		}
		cut = matches.Cut
		for _, change := range matches.Events {
			linkedChange, linkErr := changesEventProperties(change.State.Properties).linkedMergedChange(change, rez.MergedChangeLinkMergeCommit)
			if linkErr != nil {
				return nil, false, linkErr
			}
			linked = append(linked, *linkedChange)
			matched.Add(change.ID)
		}
	}

	link := rez.MergedChangeLinkMergedSincePrevious
	var mergedAfter time.Time
	if listed.PreviousDeployedAt != nil {
		mergedAfter = *listed.PreviousDeployedAt
	} else {
		link = rez.MergedChangeLinkMergedInLookback
		mergedAfter = listed.DeployedAt.Add(-changesLookbackWithoutPrevious)
	}
	// The range may hold the merge commit matches, which are listed once. Fetching as many as the limit still
	// detects a cut: the matches take the slots they leave.
	listSince := rez.ListRelatedEventsParams{
		EntityID:    *listed.RepositoryEntityID,
		Predicate:   knr.PredicateTouches,
		Kind:        changesCodeChangeEventKind,
		From:        mergedAfter,
		To:          listed.DeployedAt,
		ToInclusive: true,
		PropertyEquals: map[string]string{
			projections.CodeChangePropertyIntoDefaultBranch: strconv.FormatBool(true),
		},
		Limit: changesMaxMergedChanges,
	}
	since, sinceErr := s.graph.ListRelatedEvents(ctx, listSince)
	if sinceErr != nil {
		return nil, false, fmt.Errorf("list merged changes: %w", sinceErr)
	}
	if since.Cut {
		cut = true
	}
	for _, change := range since.Events {
		if matched.Contains(change.ID) {
			continue
		}
		if len(linked) == changesMaxMergedChanges {
			cut = true
			break
		}
		linkedChange, linkErr := changesEventProperties(change.State.Properties).linkedMergedChange(change, link)
		if linkErr != nil {
			return nil, false, linkErr
		}
		linked = append(linked, *linkedChange)
	}
	return linked, cut, nil
}

// changesEventProperties are an event entity's state properties, whose values are strings.
type changesEventProperties map[string]any

func (p changesEventProperties) get(key string) string {
	value, _ := p[key].(string)
	return value
}

// linkedMergedChange reads the merge recorded on a code change entity.
func (p changesEventProperties) linkedMergedChange(change *ent.KnowledgeEntity, link rez.MergedChangeLink) (*rez.LinkedMergedChange, error) {
	mergedAt, mergedAtErr := time.Parse(time.RFC3339Nano, p.get(projections.CodeChangePropertyMergedAt))
	if mergedAtErr != nil {
		return nil, fmt.Errorf("code change %s merge time: %w", change.ID, mergedAtErr)
	}
	number, numberErr := strconv.Atoi(p.get(projections.CodeChangePropertyNumber))
	if numberErr != nil {
		return nil, fmt.Errorf("code change %s number: %w", change.ID, numberErr)
	}
	linked := &rez.LinkedMergedChange{
		Change:            change,
		MergedAt:          mergedAt,
		MergeCommitSha:    p.get(projections.CodeChangePropertyMergeCommitSha),
		BaseRef:           p.get(projections.CodeChangePropertyBaseRef),
		IntoDefaultBranch: p.get(projections.CodeChangePropertyIntoDefaultBranch) == strconv.FormatBool(true),
		Number:            number,
		URL:               p.get(projections.CodeChangePropertyURL),
		Author:            p.get(projections.CodeChangePropertyAuthor),
		Link:              link,
	}
	return linked, nil
}
