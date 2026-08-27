package eventprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionUserProfileObserved = "user_profile_observed"
)

func (s *ProjectionService) handleUserEvent(ctx context.Context, e *projections.UserEvent) ([]rez.ProjectedEntityRef, error) {
	event := e.Event
	attributes := e.Attributes

	userObservedEvidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event),
		Assertion:   knowledgeAssertionUserProfileObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.Name,
			Description: "",
			Properties:  nil,
		},
		SubjectEntity: &ent.KnowledgeEntityRef{
			Category:        kne.CategoryActor,
			Kind:            knowledgeEntityKindUser,
			SubjectAliasRef: event.KnowledgeSubjectAliasRef(),
		},
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		subj, ingestErr := s.knowledge.IngestSubjectEvidence(ctx, event, userObservedEvidence)
		if ingestErr != nil {
			return fmt.Errorf("ingest knowledge evidence: %w", ingestErr)
		} else if subj.EntityID == nil {
			return fmt.Errorf("nil subject entity")
		}

		userId, setUserErr := s.setUserFromProjection(ctx, *subj.EntityID, attributes)
		if setUserErr != nil {
			return fmt.Errorf("create user from projection: %w", setUserErr)
		}

		projected = append(projected, rez.ProjectedEntityRef{
			Kind: knowledgeEntityKindUser,
			Id:   userId,
		})

		return nil
	})
}

func (s *ProjectionService) setUserFromProjection(ctx context.Context, knowledgeEntityId uuid.UUID, attrs projections.UserSubjectAttributes) (uuid.UUID, error) {
	linkedPred := user.KnowledgeEntityID(knowledgeEntityId)
	if attrs.Email != "" {
		linkedPred = user.Or(linkedPred, user.Email(attrs.Email))
	}
	queryLinked := s.db.Client(ctx).User.Query().
		Where(linkedPred)
	linked, linkedErr := queryLinked.Only(ctx)
	if linkedErr != nil && !ent.IsNotFound(linkedErr) {
		return uuid.Nil, fmt.Errorf("query user by knowledge entity: %w", linkedErr)
	}

	var userID uuid.UUID
	if linked != nil {
		userID = linked.ID
	}

	usr, setErr := s.users.Set(ctx, userID, func(m *ent.UserMutation) {
		m.SetKnowledgeEntityID(knowledgeEntityId)
		m.SetName(attrs.Name)
		m.SetEmail(attrs.Email)
		m.SetChatID(attrs.ChatId)
		m.SetTimezone(attrs.Timezone)
	})
	if setErr != nil {
		return uuid.Nil, fmt.Errorf("save user: %w", setErr)
	}
	return usr.ID, nil
}
