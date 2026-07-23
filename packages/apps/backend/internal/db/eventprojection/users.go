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

func (s *ProjectionService) handleUserEvent(ctx context.Context, event *projections.UserEvent) ([]rez.ProjectedEntityRef, error) {
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
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		knSubj, knowledgeErr := s.knowledge.IngestDomainEntityEvidence(ctx, event.Event, userObservedEvidence)
		if knowledgeErr != nil {
			return fmt.Errorf("resolve user knowledge entity: %w", knowledgeErr)
		}

		queryLinked := tx.User.Query().
			Where(user.Or(user.KnowledgeEntityID(knSubj.EntityID), user.Email(attributes.Email)))
		linked, linkedErr := queryLinked.Only(ctx)
		if linkedErr != nil && !ent.IsNotFound(linkedErr) {
			return fmt.Errorf("query linked user: %w", linkedErr)
		}

		var userID uuid.UUID
		if linked != nil {
			userID = linked.ID
		}

		usr, setErr := s.users.Set(ctx, userID, func(m *ent.UserMutation) {
			m.SetKnowledgeEntityID(knSubj.EntityID)
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
	})
}
