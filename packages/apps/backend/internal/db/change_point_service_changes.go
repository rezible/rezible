package db

import (
	"context"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	changesMaxDeployments          = 20
	changesMaxMergedChanges        = 20
	changesLookbackWithoutPrevious = 7 * 24 * time.Hour
)

func (s *ChangePointService) ListServiceChanges(ctx context.Context, params rez.ListServiceChangesParams) (*rez.ServiceChanges, error) {
	if !params.End.After(params.Start) {
		return nil, fmt.Errorf("%w: end must be after start", errs.ErrInvalidInput)
	}
	if params.End.Sub(params.Start) > rez.MaxServiceChangesWindow {
		return nil, fmt.Errorf("%w: the window is longer than the 7 day limit", errs.ErrInvalidInput)
	}
	environment := ""
	if params.Environment != "" {
		environment = projections.NormalizeServiceName(params.Environment)
		if environment == "" {
			return nil, fmt.Errorf("%w: environment %q has no name once normalized", errs.ErrInvalidInput, params.Environment)
		}
	}

	client := s.db.Client(ctx)
	serviceID, canonicalErr := ent.CanonicalKnowledgeEntityID(ctx, client, params.ServiceEntityID)
	if canonicalErr != nil {
		return nil, fmt.Errorf("resolve service redirect: %w", canonicalErr)
	}
	queryService := client.KnowledgeEntity.Query().
		Where(kne.ID(serviceID), currentKnowledgeEntity())
	serviceEntity, serviceErr := queryService.Only(ctx)
	if serviceErr != nil && !ent.IsNotFound(serviceErr) {
		return nil, fmt.Errorf("query service entity: %w", serviceErr)
	}
	if serviceEntity == nil || !knowledgegraph.IsService(serviceEntity.Category, serviceEntity.Kind) {
		return nil, fmt.Errorf("%w: %s is not a known service", errs.ErrInvalidInput, params.ServiceEntityID)
	}

	listDeployments := rez.ListChangePointsParams{
		EntityID:      serviceID,
		Roles:         []string{projections.AssertionDeploymentService},
		Kind:          projections.KindDeployment,
		From:          params.Start,
		FromInclusive: true,
		To:            params.End,
		ToInclusive:   true,
		Limit:         changesMaxDeployments,
	}
	if environment != "" {
		listDeployments.AttributeEquals = []rez.ChangePointAttribute{{Path: projections.DeploymentEnvironmentName, Value: environment}}
	}
	deployments, deploymentsErr := s.ListChangePoints(ctx, listDeployments)
	if deploymentsErr != nil {
		return nil, fmt.Errorf("list deployments: %w", deploymentsErr)
	}

	result := &rez.ServiceChanges{}
	if deployments.Cut {
		limit := fmt.Sprintf("more than %d deployments in the window; listing the %d most recent", changesMaxDeployments, changesMaxDeployments)
		result.Limits = append(result.Limits, limit)
	}
	for _, d := range deployments.ChangePoints {
		listed, listErr := s.listDeploymentChanges(ctx, serviceID, d)
		if listErr != nil {
			return nil, fmt.Errorf("deployment %s: %w", d.ChangePoint.ID, listErr)
		}
		result.Deployments = append(result.Deployments, *listed)
	}
	return result, nil
}

func (s *ChangePointService) listDeploymentChanges(ctx context.Context, serviceEntityID uuid.UUID, d rez.ChangePoint) (*rez.ServiceDeployment, error) {
	decoded, decodeErr := projections.DecodeDeploymentEvent(d.ChangePoint.Edges.LatestEvent)
	if decodeErr != nil {
		return nil, fmt.Errorf("decode deployment: %w", decodeErr)
	}
	attributes := decoded.Attributes
	listed := &rez.ServiceDeployment{
		Deployment:  d,
		DeployedAt:  d.ChangePoint.OccurredAt,
		Environment: attributes.Environment.Name,
		Status:      attributes.Status,
		Sha:         attributes.Sha,
		Version:     attributes.Version,
		URL:         attributes.URL,
	}
	repositoryFullName := ""
	if attributes.Repository != nil {
		repositoryFullName = attributes.Repository.LinkingAttributes[projections.LinkingAttributeRepositoryFullName]
	}
	listPrevious := rez.ListChangePointsParams{
		EntityID: serviceEntityID,
		Roles:    []string{projections.AssertionDeploymentService},
		Kind:     projections.KindDeployment,
		To:       listed.DeployedAt,
		AttributeEquals: []rez.ChangePointAttribute{
			{Path: projections.DeploymentStatus, Value: projections.DeploymentStatusSucceeded},
			{Path: projections.DeploymentEnvironmentName, Value: listed.Environment},
		},
		Limit: 1,
	}
	if repositoryFullName != "" {
		listPrevious.AttributeEquals = append(listPrevious.AttributeEquals, rez.ChangePointAttribute{Path: projections.DeploymentRepositoryFullName, Value: repositoryFullName})
	} else {
		listPrevious.AttributeAbsent = []projections.AttributePath{projections.DeploymentRepositoryFullName}
	}
	previous, previousErr := s.ListChangePoints(ctx, listPrevious)
	if previousErr != nil {
		return nil, fmt.Errorf("list previous deployment: %w", previousErr)
	}
	if len(previous.ChangePoints) > 0 {
		listed.PreviousDeployedAt = &previous.ChangePoints[0].ChangePoint.OccurredAt
	}
	var repositoryEntityID *uuid.UUID
	for _, involved := range d.Involved {
		if involved.Role == projections.AssertionDeploymentRepository {
			repositoryEntityID = involved.Edges.SubjectAlias.EntityID
			break
		}
	}
	if repositoryEntityID == nil {
		return listed, nil
	}
	repository, repositoryErr := s.queryDeploymentRepository(ctx, *repositoryEntityID)
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

// queryDeploymentRepository loads the involved repository with aliases observed by a code forge.
func (s *ChangePointService) queryDeploymentRepository(ctx context.Context, repositoryID uuid.UUID) (*ent.KnowledgeEntity, error) {
	return s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(repositoryID)).
		WithAliases(func(q *ent.KnowledgeSubjectAliasQuery) {
			q.Where(ksa.HasEvidenceWith(ke.HasEventWith(ne.KindIn(projections.KindCodeForge, projections.KindCodeChange))))
		}).
		Only(ctx)
}

// linkMergedChanges lists the merged changes of the deployment's repository linked to it: merge commit
// matches first, then merges into the default branch since the previous deployment, or within the lookback
// when there is none. Merge commit matches are always kept; the rest fill the remaining slots. It reports
// whether the limit cut the list.
func (s *ChangePointService) linkMergedChanges(ctx context.Context, listed *rez.ServiceDeployment) ([]rez.LinkedMergedChange, bool, error) {
	var linked []rez.LinkedMergedChange
	matched := mapset.NewSet[uuid.UUID]()
	cut := false

	if listed.Sha != "" {
		listMatches := rez.ListChangePointsParams{
			EntityID:        *listed.RepositoryEntityID,
			Roles:           []string{projections.AssertionChangeRepository},
			Kind:            projections.KindCodeChange,
			AttributeEquals: []rez.ChangePointAttribute{{Path: projections.CodeChangeMergeCommitSha, Value: listed.Sha}},
			Limit:           changesMaxMergedChanges,
		}
		matches, matchesErr := s.ListChangePoints(ctx, listMatches)
		if matchesErr != nil {
			return nil, false, fmt.Errorf("list merge commit matches: %w", matchesErr)
		}
		cut = matches.Cut
		for _, change := range matches.ChangePoints {
			linkedChange, linkErr := s.linkedMergedChange(change, rez.MergedChangeLinkMergeCommit)
			if linkErr != nil {
				return nil, false, linkErr
			}
			linked = append(linked, *linkedChange)
			matched.Add(change.ChangePoint.ID)
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
	listSince := rez.ListChangePointsParams{
		EntityID:        *listed.RepositoryEntityID,
		Roles:           []string{projections.AssertionChangeRepository},
		Kind:            projections.KindCodeChange,
		From:            mergedAfter,
		To:              listed.DeployedAt,
		ToInclusive:     true,
		AttributeEquals: []rez.ChangePointAttribute{{Path: projections.CodeChangeIntoDefaultBranch, Value: "true"}},
		Limit:           changesMaxMergedChanges,
	}
	since, sinceErr := s.ListChangePoints(ctx, listSince)
	if sinceErr != nil {
		return nil, false, fmt.Errorf("list merged changes: %w", sinceErr)
	}
	if since.Cut {
		cut = true
	}
	for _, change := range since.ChangePoints {
		if matched.Contains(change.ChangePoint.ID) {
			continue
		}
		if len(linked) == changesMaxMergedChanges {
			cut = true
			break
		}
		linkedChange, linkErr := s.linkedMergedChange(change, link)
		if linkErr != nil {
			return nil, false, linkErr
		}
		linked = append(linked, *linkedChange)
	}
	return linked, cut, nil
}

func (s *ChangePointService) linkedMergedChange(change rez.ChangePoint, link rez.MergedChangeLink) (*rez.LinkedMergedChange, error) {
	decoded, decodeErr := projections.DecodeCodeChangeEvent(change.ChangePoint.Edges.LatestEvent)
	if decodeErr != nil {
		return nil, fmt.Errorf("decode merged change: %w", decodeErr)
	}
	merge := decoded.Attributes.Merge
	if merge == nil {
		return nil, fmt.Errorf("change point %s has no merge", change.ChangePoint.ID)
	}
	return &rez.LinkedMergedChange{
		Change:            change,
		Title:             decoded.Attributes.DisplayName,
		MergedAt:          change.ChangePoint.OccurredAt,
		MergeCommitSha:    merge.MergeCommitSha,
		BaseRef:           merge.BaseRef,
		IntoDefaultBranch: merge.IntoDefaultBranch,
		Number:            merge.Number,
		URL:               merge.URL,
		Author:            merge.AuthorLogin,
		Link:              link,
	}, nil
}
