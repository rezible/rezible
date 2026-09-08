package eventprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	entschema "github.com/rezible/rezible/ent/schema"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/teammembership"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionTeamObserved           = "team_profile_observed"
	knowledgeAssertionTeamMembershipObserved = "team_membership_observed"
)

func (s *ProjectionService) handleTeamEvent(ctx context.Context, e *projections.TeamEvent) ([]rez.ProjectedEntityRef, error) {
	attrs := e.Attributes
	teamResourceRef := rez.ProviderResourceRef{
		Provider:          e.Event.Provider,
		ProviderNamespace: e.Event.ProviderNamespace,
		ResourceRef:       e.Event.ProviderResourceRef,
	}
	teamEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryActor,
		Kind:                knowledgeEntityKindTeam,
		ProviderResourceRef: teamResourceRef,
	}
	evidence := rez.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(e.Event),
		Assertion:   knowledgeAssertionTeamObserved,
		EffectiveAt: e.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Name,
			Properties: map[string]any{
				"slug":            attrs.Slug,
				"chat_channel_id": attrs.ChatChannelId,
			},
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &teamEntityRef},
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		subj, ingestErr := s.ingestSubjectEvidence(ctx, e.Event, evidence)
		if ingestErr != nil {
			return fmt.Errorf("ingest knowledge evidence: %w", ingestErr)
		} else if subj.EntityID == nil {
			return fmt.Errorf("nil subject entity")
		}

		teamId, saveTeamErr := s.setTeamFromProjection(ctx, *subj.EntityID, attrs)
		if saveTeamErr != nil {
			return fmt.Errorf("create team from projection: %w", saveTeamErr)
		}

		projected = append(projected, rez.ProjectedEntityRef{
			Kind: knowledgeEntityKindTeam,
			Id:   teamId,
		})
		return nil
	})
}

func (s *ProjectionService) setTeamFromProjection(ctx context.Context, knowledgeEntityId uuid.UUID, attrs projections.TeamEventAttributes) (uuid.UUID, error) {
	client := s.db.Client(ctx)
	lookupTeam := client.Team.Query().
		Where(team.Or(team.KnowledgeEntityID(knowledgeEntityId), team.Slug(attrs.Slug)))
	existing, queryErr := lookupTeam.Only(entschema.IncludeArchived(ctx))
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query team: %w", queryErr)
	}

	var upsert ent.EntityMutator[*ent.Team, *ent.TeamMutation]
	if existing == nil {
		upsert = client.Team.Create()
	} else {
		upsert = existing.Update().ClearArchiveTime()
	}
	m := upsert.Mutation()
	m.SetKnowledgeEntityID(knowledgeEntityId)
	m.SetSlug(attrs.Slug)
	m.SetName(attrs.Name)
	m.SetChatChannelID(attrs.ChatChannelId)

	saved, saveErr := upsert.Save(ctx)
	if saveErr != nil {
		return uuid.Nil, fmt.Errorf("save team: %w", saveErr)
	}
	return saved.ID, nil
}

func (s *ProjectionService) handleTeamMembershipEvent(ctx context.Context, e *projections.TeamMembershipEvent) ([]rez.ProjectedEntityRef, error) {
	attrs := e.Attributes
	event := e.Event

	kind := projectionEvidenceKind(event)

	userEntity := rez.KnowledgeEntityRef{
		Category:            kne.CategoryActor,
		Kind:                knowledgeEntityKindUser,
		ProviderResourceRef: attrs.User.ProviderResourceRef,
		LinkingAttributes:   projections.UserEntityLinkingAttributes{Email: attrs.User.Email},
	}

	teamEntity := rez.KnowledgeEntityRef{
		Category:            kne.CategoryActor,
		Kind:                knowledgeEntityKindTeam,
		ProviderResourceRef: attrs.Team.ProviderResourceRef,
	}
	userEvidenceRef := rez.KnowledgeEvidenceRef{
		Kind:        kind,
		Assertion:   knowledgeAssertionUserProfileObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.User.Name,
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &userEntity},
	}
	teamEvidenceRef := rez.KnowledgeEvidenceRef{
		Kind:        kind,
		Assertion:   knowledgeAssertionTeamObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Team.Name,
			Properties: map[string]any{
				"slug":            attrs.Team.Slug,
				"chat_channel_id": attrs.Team.ChatChannelId,
			},
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &teamEntity},
	}

	membershipResourceRef := rez.ProviderResourceRef{
		Provider:          event.Provider,
		ProviderNamespace: event.ProviderNamespace,
		ResourceRef:       event.ProviderResourceRef,
	}
	membershipRelationship := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateMemberOf,
		ProviderResourceRef: membershipResourceRef,
		Source:              userEntity,
		Target:              teamEntity,
	}
	membershipEvidenceRef := rez.KnowledgeEvidenceRef{
		Kind:        kind,
		Assertion:   knowledgeAssertionTeamMembershipObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Member of " + attrs.Team.Name,
			Properties:  map[string]any{"role": attrs.Role},
		},
		Subject: rez.KnowledgeSubjectRef{Relationship: &membershipRelationship},
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		subj, ingestErr := s.ingestSubjectEvidence(ctx, event, membershipEvidenceRef, userEvidenceRef, teamEvidenceRef)
		if ingestErr != nil {
			return fmt.Errorf("ingest knowledge evidence: %w", ingestErr)
		} else if subj.RelationshipID == nil {
			return fmt.Errorf("nil subject relationship")
		}

		membershipId, membershipErr := s.setTeamMembershipFromProjection(ctx, *subj.RelationshipID, attrs)
		if membershipErr != nil {
			return fmt.Errorf("set membership: %w", membershipErr)
		}

		projected = append(projected, rez.ProjectedEntityRef{
			Kind: "team_membership",
			Id:   membershipId,
		})
		return nil
	})
}

func (s *ProjectionService) setTeamMembershipFromProjection(ctx context.Context, relId uuid.UUID, attrs projections.TeamMembershipEventAttributes) (uuid.UUID, error) {
	rel, relErr := s.knowledge.GetRelationship(ctx, relId)
	if relErr != nil {
		return uuid.Nil, fmt.Errorf("get knowledge relationship: %w", relErr)
	}

	userEntityId := rel.SourceEntityID
	teamEntityId := rel.TargetEntityID
	if rel.Edges.TargetEntity != nil && rel.Edges.TargetEntity.Kind == knowledgeEntityKindUser {
		userEntityId = rel.TargetEntityID
		teamEntityId = rel.SourceEntityID
	}

	userAttributes := projections.UserEventAttributes{
		Name:     attrs.User.Name,
		Email:    attrs.User.Email,
		ChatId:   attrs.User.ChatId,
		Timezone: attrs.User.Timezone,
	}
	userId, userErr := s.setUserFromProjection(ctx, userEntityId, userAttributes)
	if userErr != nil {
		return uuid.Nil, fmt.Errorf("save membership user: %w", userErr)
	}

	teamAttributes := projections.TeamEventAttributes{
		Name:          attrs.Team.Name,
		Slug:          attrs.Team.Slug,
		ChatChannelId: attrs.Team.ChatChannelId,
	}
	teamId, teamErr := s.setTeamFromProjection(ctx, teamEntityId, teamAttributes)
	if teamErr != nil {
		return uuid.Nil, fmt.Errorf("save membership team: %w", teamErr)
	}

	role := teammembership.Role(attrs.Role)

	client := s.db.Client(ctx)
	lookupExisting := client.TeamMembership.Query().
		Where(teammembership.TeamID(teamId), teammembership.UserID(userId))
	existing, queryErr := lookupExisting.Only(entschema.IncludeArchived(ctx))
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query existing: %w", queryErr)
	}

	var upsert ent.EntityMutator[*ent.TeamMembership, *ent.TeamMembershipMutation]
	if existing == nil {
		upsert = client.TeamMembership.Create()
	} else {
		if existing.Role == role {
			return existing.ID, nil
		}
		upsert = existing.Update()
	}

	//if e.Event.Kind == ne.KindDeleted {
	//	if membership != nil {
	//		if deleteErr := tx.TeamMembership.DeleteOne(membership).Exec(ctx); deleteErr != nil {
	//			return uuid.Nil, fmt.Errorf("delete membership: %w", deleteErr)
	//		}
	//	}
	//	return nil
	//}

	m := upsert.Mutation()
	m.SetTeamID(teamId)
	m.SetUserID(userId)
	m.SetRole(role)

	saved, saveErr := upsert.Save(ctx)
	if saveErr != nil {
		return uuid.Nil, fmt.Errorf("save team: %w", saveErr)
	}
	return saved.ID, nil
}
