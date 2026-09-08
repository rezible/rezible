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

	userResourceRef := rez.ProviderResourceRef{
		Provider:          event.Provider,
		ProviderNamespace: event.ProviderNamespace,
		ResourceRef:       event.ProviderResourceRef,
	}
	userEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryActor,
		Kind:                knowledgeEntityKindUser,
		ProviderResourceRef: userResourceRef,
		LinkingAttributes:   projections.UserEntityLinkingAttributes{Email: attributes.Email},
	}
	userObservedEvidence := rez.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event),
		Assertion:   knowledgeAssertionUserProfileObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.Name,
			Description: "",
			Properties:  nil,
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &userEntityRef},
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		subj, ingestErr := s.ingestSubjectEvidence(ctx, event, userObservedEvidence)
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

func (s *ProjectionService) setUserFromProjection(ctx context.Context, knowledgeEntityId uuid.UUID, attrs projections.UserEventAttributes) (uuid.UUID, error) {
	client := s.db.Client(ctx)
	linked, linkedErr := client.User.Query().Where(user.KnowledgeEntityID(knowledgeEntityId)).Only(ctx)
	if linkedErr != nil && !ent.IsNotFound(linkedErr) {
		return uuid.Nil, fmt.Errorf("query user by knowledge entity: %w", linkedErr)
	}
	if ent.IsNotFound(linkedErr) {
		linked = nil
	}

	if attrs.Email != "" {
		emailOwner, emailErr := client.User.Query().Where(user.Email(attrs.Email)).Only(ctx)
		if emailErr != nil && !ent.IsNotFound(emailErr) {
			return uuid.Nil, fmt.Errorf("query user by email: %w", emailErr)
		}
		if ent.IsNotFound(emailErr) {
			emailOwner = nil
		}
		if emailOwner != nil {
			if linked != nil && emailOwner.ID != linked.ID {
				return uuid.Nil, fmt.Errorf("%w: email is already assigned to another user", rez.ErrConflict)
			}
			if emailOwner.KnowledgeEntityID != nil && *emailOwner.KnowledgeEntityID != knowledgeEntityId {
				return uuid.Nil, fmt.Errorf("%w: email user is linked to another knowledge entity", rez.ErrConflict)
			}
			if linked == nil {
				linked = emailOwner
			}
		}
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
