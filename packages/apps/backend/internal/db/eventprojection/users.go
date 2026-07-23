package eventprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeEntityKindUser               = "user"
	knowledgeAssertionUserProfileObserved = "user_profile_observed"
)

func (s *ProjectionService) handleUserEventProjection(ctx context.Context, normalizedEvent *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	event, decodeErr := projections.DecodeUserEvent(normalizedEvent)
	if decodeErr != nil {
		return nil, fmt.Errorf("invalid user event: %w", decodeErr)
	}
	attributes := event.Attributes

	userSubjectRef := event.Event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "User")
	userSubjectRef.SubjectEntityRef = &ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindUser,
		Reference:   attributes.Email,
		DisplayName: attributes.Name,
	}
	userObservedEvidence := ent.KnowledgeEvidenceRef{
		Kind:            projectionEvidenceKind(event.Event),
		Assertion:       knowledgeAssertionUserProfileObserved,
		EffectiveAt:     event.Event.OccurredAt,
		SubjectAliasRef: userSubjectRef,
	}

	var projected []rez.ProjectedEntityRef
	projectTxFn := func(txCtx context.Context, tx *ent.Client) error {
		knowledgeEntityID, knowledgeErr := s.ingestDomainEntityEvidence(txCtx, event.Event, userObservedEvidence)
		if knowledgeErr != nil {
			return fmt.Errorf("resolve user knowledge entity: %w", knowledgeErr)
		}

		queryLinked := tx.User.Query().
			Where(user.Or(user.KnowledgeEntityID(knowledgeEntityID), user.Email(attributes.Email)))
		linked, linkedErr := queryLinked.Only(txCtx)
		if linkedErr != nil && !ent.IsNotFound(linkedErr) {
			return fmt.Errorf("query linked user: %w", linkedErr)
		}

		var userID uuid.UUID
		if linked != nil {
			userID = linked.ID
		}

		usr, setErr := s.users.Set(txCtx, userID, func(m *ent.UserMutation) {
			m.SetKnowledgeEntityID(knowledgeEntityID)
			m.SetName(attributes.Name)
			m.SetEmail(attributes.Email)
			m.SetChatID(attributes.ChatId)
			m.SetTimezone(attributes.Timezone)
		})
		if setErr != nil {
			return fmt.Errorf("save user: %w", setErr)
		}

		projected = append(projected, rez.ProjectedEntityRef{Kind: knowledgeEntityKindUser, Id: usr.ID})

		return nil
	}
	return projected, s.db.WithTx(ctx, projectTxFn)
}
