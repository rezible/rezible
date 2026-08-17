package eventprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	entschema "github.com/rezible/rezible/ent/schema"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/teammembership"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionTeamObserved = "team_profile_observed"
	knowledgeAssertionMembership   = "team_membership_observed"
)

func (s *ProjectionService) handleTeamEvent(ctx context.Context, e *projections.TeamEvent) ([]rez.ProjectedEntityRef, error) {
	attrs := e.Attributes
	evidence := ent.KnowledgeEvidenceRef{
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
		SubjectEntity: &ent.KnowledgeEntityRef{
			Kind:    kne.KindActor,
			Subkind: knowledgeEntitySubkindTeam,
			Alias:   e.Event.KnowledgeAliasRef(),
		},
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		entity, ingestErr := s.knowledge.IngestEntityEvidence(ctx, e.Event, evidence)
		if ingestErr != nil {
			return fmt.Errorf("ingest team evidence: %w", ingestErr)
		}

		lookupTeam := tx.Team.Query().
			Where(team.Or(team.KnowledgeEntityID(entity.ID), team.Slug(attrs.Slug)))
		existing, queryErr := lookupTeam.Only(entschema.IncludeArchived(ctx))
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return fmt.Errorf("query team: %w", queryErr)
		}

		if e.Event.Kind == ne.KindDeleted {
			if existing != nil {
				if deleteErr := tx.Team.DeleteOne(existing).Exec(ctx); deleteErr != nil {
					return fmt.Errorf("archive team: %w", deleteErr)
				}
			}
			return nil
		}

		var upsert ent.EntityMutator[*ent.Team, *ent.TeamMutation]
		if existing == nil {
			upsert = tx.Team.Create()
		} else {
			upsert = existing.Update().ClearArchiveTime()
		}
		m := upsert.Mutation()
		m.SetKnowledgeEntityID(entity.ID)
		m.SetSlug(attrs.Slug)
		m.SetName(attrs.Name)
		m.SetChatChannelID(attrs.ChatChannelId)

		saved, saveErr := upsert.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save team: %w", saveErr)
		}
		projected = append(projected, rez.ProjectedEntityRef{Kind: knowledgeEntitySubkindTeam, Id: saved.ID})
		return nil
	})
}

func (s *ProjectionService) handleTeamMembershipEvent(ctx context.Context, event *projections.TeamMembershipEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes
	userAlias := ent.KnowledgeAliasRef{
		Provider:           event.Event.Provider,
		ProviderSource:     "users",
		ProviderSubjectRef: attributes.UserExternalRef,
	}
	teamAlias := ent.KnowledgeAliasRef{
		Provider:           event.Event.Provider,
		ProviderSource:     "teams",
		ProviderSubjectRef: attributes.Team.ExternalRef,
	}
	kind := projectionEvidenceKind(event.Event)
	refs := []ent.KnowledgeEvidenceRef{
		{
			Kind:        kind,
			Assertion:   knowledgeAssertionUserProfileObserved,
			EffectiveAt: event.Event.OccurredAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: attributes.User.Name,
			},
			SubjectEntity: &ent.KnowledgeEntityRef{Kind: kne.KindActor, Subkind: knowledgeEntitySubkindUser, Alias: userAlias},
		},
		{
			Kind:        kind,
			Assertion:   knowledgeAssertionTeamObserved,
			EffectiveAt: event.Event.OccurredAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: attributes.Team.Name,
			},
			SubjectEntity: &ent.KnowledgeEntityRef{Kind: kne.KindActor, Subkind: knowledgeEntitySubkindTeam, Alias: teamAlias},
		},
		{
			Kind:        kind,
			Assertion:   knowledgeAssertionMembership,
			EffectiveAt: event.Event.OccurredAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: "Member of " + attributes.Team.Name,
				Properties:  map[string]any{"role": attributes.Role},
			},
			SubjectRelationship: &ent.KnowledgeRelationshipRef{
				Kind:    knr.KindParticipatesIn,
				Subkind: knowledgeRelationshipSubkindMemberOf,
				Alias:   event.Event.KnowledgeAliasRef(),
				Source:  ent.KnowledgeEntityRef{Kind: kne.KindActor, Subkind: knowledgeEntitySubkindUser, Alias: userAlias},
				Target:  ent.KnowledgeEntityRef{Kind: kne.KindActor, Subkind: knowledgeEntitySubkindTeam, Alias: teamAlias},
			},
		},
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		aliases, ingestErr := s.knowledge.IngestEvidenceBulk(ctx, event.Event, refs...)
		if ingestErr != nil {
			return fmt.Errorf("ingest membership evidence: %w", ingestErr)
		}
		if len(aliases) != 3 || aliases[0].EntityID == nil || aliases[1].EntityID == nil {
			return fmt.Errorf("membership evidence did not resolve its entities")
		}

		usr, userErr := tx.User.Query().
			Where(user.Or(user.KnowledgeEntityID(*aliases[0].EntityID), user.Email(attributes.User.Email))).
			Only(ctx)
		if userErr != nil && !ent.IsNotFound(userErr) {
			return fmt.Errorf("query membership user: %w", userErr)
		}
		userID := uuid.Nil
		if usr == nil {
			usr, userErr = tx.User.Create().
				SetKnowledgeEntityID(*aliases[0].EntityID).
				SetName(attributes.User.Name).
				SetEmail(attributes.User.Email).
				SetChatID(attributes.User.ChatId).
				SetTimezone(attributes.User.Timezone).
				Save(ctx)
		} else {
			usr, userErr = usr.Update().
				SetKnowledgeEntityID(*aliases[0].EntityID).
				SetName(attributes.User.Name).
				SetEmail(attributes.User.Email).
				SetChatID(attributes.User.ChatId).
				SetTimezone(attributes.User.Timezone).
				Save(ctx)
		}
		if userErr != nil {
			return fmt.Errorf("save membership user: %w", userErr)
		}
		userID = usr.ID

		teamRecord, teamErr := tx.Team.Query().
			Where(team.Or(team.KnowledgeEntityID(*aliases[1].EntityID), team.Slug(attributes.Team.Slug))).
			Only(entschema.IncludeArchived(ctx))
		if teamErr != nil && !ent.IsNotFound(teamErr) {
			return fmt.Errorf("query membership team: %w", teamErr)
		}
		if teamRecord == nil {
			teamRecord, teamErr = tx.Team.Create().
				SetKnowledgeEntityID(*aliases[1].EntityID).
				SetSlug(attributes.Team.Slug).
				SetName(attributes.Team.Name).
				SetChatChannelID(attributes.Team.ChatChannelId).
				Save(ctx)
		} else {
			teamRecord, teamErr = teamRecord.Update().
				SetKnowledgeEntityID(*aliases[1].EntityID).
				SetSlug(attributes.Team.Slug).
				SetName(attributes.Team.Name).
				SetChatChannelID(attributes.Team.ChatChannelId).
				ClearArchiveTime().
				Save(ctx)
		}
		if teamErr != nil {
			return fmt.Errorf("save membership team: %w", teamErr)
		}

		membership, membershipErr := tx.TeamMembership.Query().
			Where(teammembership.TeamID(teamRecord.ID), teammembership.UserID(userID)).
			Only(ctx)
		if membershipErr != nil && !ent.IsNotFound(membershipErr) {
			return fmt.Errorf("query membership: %w", membershipErr)
		}
		if event.Event.Kind == ne.KindDeleted {
			if membership != nil {
				if deleteErr := tx.TeamMembership.DeleteOne(membership).Exec(ctx); deleteErr != nil {
					return fmt.Errorf("delete membership: %w", deleteErr)
				}
			}
			return nil
		}
		role := teammembership.Role(attributes.Role)
		if membership == nil {
			membership, membershipErr = tx.TeamMembership.Create().
				SetTeamID(teamRecord.ID).
				SetUserID(userID).
				SetRole(role).
				Save(ctx)
		} else {
			membership, membershipErr = membership.Update().SetRole(role).Save(ctx)
		}
		if membershipErr != nil {
			return fmt.Errorf("save membership: %w", membershipErr)
		}
		projected = append(projected, rez.ProjectedEntityRef{Kind: "team_membership", Id: membership.ID})
		return nil
	})
}
