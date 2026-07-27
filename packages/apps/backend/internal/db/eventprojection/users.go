package eventprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeEntityKindUser               = "user"
	knowledgeAssertionUserProfileObserved = "user_profile_observed"
)

func (s *ProjectionService) handleUserEvent(ctx context.Context, event *projections.UserEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes

	userObservedEvidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionUserProfileObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.Name,
			Description: "",
			Properties:  nil,
		},
		SubjectEntity: &ent.KnowledgeEntityRef{
			Kind:  knowledgeEntityKindUser,
			Alias: event.Event.KnowledgeAliasRef(),
		},
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		knSubj, knowledgeErr := s.knowledge.IngestEntityEvidence(ctx, event.Event, userObservedEvidence)
		if knowledgeErr != nil {
			return fmt.Errorf("resolve user knowledge entity: %w", knowledgeErr)
		}

		queryLinked := tx.User.Query().
			Where(user.Or(user.KnowledgeEntityID(knSubj.ID), user.Email(attributes.Email)))
		linked, linkedErr := queryLinked.Only(ctx)
		if linkedErr != nil && !ent.IsNotFound(linkedErr) {
			return fmt.Errorf("query linked user: %w", linkedErr)
		}

		var userID uuid.UUID
		if linked != nil {
			userID = linked.ID
		}

		usr, setErr := s.users.Set(ctx, userID, func(m *ent.UserMutation) {
			m.SetKnowledgeEntityID(knSubj.ID)
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
