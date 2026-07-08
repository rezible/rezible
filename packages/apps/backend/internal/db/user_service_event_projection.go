package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	assertionUserProfileObserved = "user_profile_observed"
	knowledgeKindUser            = "user"
)

func (s *UserService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) ([]rez.ProjectedDomainEntityRef, error) {
	if projections.SubjectKindUser.Matches(event) {
		decoded, eventErr := projections.DecodeUserEvent(event)
		if eventErr != nil || decoded == nil {
			return nil, fmt.Errorf("invalid event: %w", eventErr)
		}
		return s.handleUserEventProjection(ctx, decoded)
	}
	return nil, nil
}

func (s *UserService) handleUserEventProjection(ctx context.Context, ue *projections.UserEvent) ([]rez.ProjectedDomainEntityRef, error) {
	attrs := ue.Attributes

	userSubjectAliasRef := ue.Event.MakeSubjectAliasRef(ksa.SubjectKindEntity)
	userObservedEvidence := rez.ProjectedKnowledgeEvidence{
		Kind:        ke.EvidenceKindObserved,
		Assertion:   assertionUserProfileObserved,
		EffectiveAt: ue.Event.OccurredAt,
		SubjectAlias: rez.ProjectedKnowledgeEvidenceSubjectAlias{
			Description: "User Profile",
			AliasRef:    userSubjectAliasRef,
			SubjectEntityRef: &ent.KnowledgeEntityRef{
				Kind:        knowledgeKindUser,
				Reference:   attrs.Email,
				DisplayName: attrs.Name,
				Description: "",
			},
		},
	}
	knowledgeEvidence := []rez.ProjectedKnowledgeEvidence{userObservedEvidence}

	var projEnts []rez.ProjectedDomainEntityRef
	return projEnts, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		evidenceErr := s.knowledge.IngestProjectedEventEvidence(ctx, ue.Event, knowledgeEvidence)
		if evidenceErr != nil {
			return fmt.Errorf("user knowledge entity: %w", evidenceErr)
		}

		keId, knowledgeErr := s.knowledge.LookupEntityIdByAliasRefs(ctx, userSubjectAliasRef)
		if knowledgeErr != nil {
			return fmt.Errorf("resolve user knowledge entity: %w", knowledgeErr)
		}

		queryLinked := s.db.Client(ctx).User.Query().
			Where(user.Or(user.KnowledgeEntityID(keId), user.Email(attrs.Email)))
		linked, linkedErr := queryLinked.Only(ctx)
		if linkedErr != nil && !ent.IsNotFound(linkedErr) {
			return fmt.Errorf("query linked user: %w", linkedErr)
		}

		var userId uuid.UUID
		if linked != nil {
			userId = linked.ID
		}

		usr, setErr := s.Set(ctx, userId, func(m *ent.UserMutation) {
			m.SetKnowledgeEntityID(keId)
			m.SetName(attrs.Name)
			m.SetEmail(attrs.Email)
			m.SetChatID(attrs.ChatId)
			m.SetTimezone(attrs.Timezone)
		})
		if setErr != nil {
			return fmt.Errorf("save user: %w", setErr)
		}

		projEnts = append(projEnts, rez.ProjectedDomainEntityRef{Kind: "user", Id: usr.ID})

		return nil
	})
}
