package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	assertionUserProfileObserved = "user_profile_observed"
	knowledgeKindUser            = "user"
)

func (s *UserService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) (map[string][]uuid.UUID, error) {
	if projections.SubjectKindUser.Matches(event) {
		decoded, eventErr := projections.DecodeUserEvent(event)
		if eventErr != nil || decoded == nil {
			return nil, fmt.Errorf("invalid event: %w", eventErr)
		}
		return s.handleUserEventProjection(ctx, decoded)
	}
	return nil, nil
}

func (s *UserService) handleUserEventProjection(ctx context.Context, ue *projections.UserEvent) (map[string][]uuid.UUID, error) {
	attrs := ue.Attributes
	projKnowledgeEntity := rez.ProjectedKnowledgeEntity{
		EvidenceAssertion: assertionUserProfileObserved,
		Kind:              knowledgeKindUser,
		DisplayName:       attrs.Name,
		AliasRefs: []ent.KnowledgeEntityAliasRef{
			{Provider: ue.Event.Provider, ProviderSubjectRef: ue.Event.ProviderSubjectRef},
		},
	}
	keId, knowledgeErr := s.knowledge.ResolveProjectedEntity(ctx, ue.Event, projKnowledgeEntity)
	if knowledgeErr != nil {
		return nil, fmt.Errorf("resolve user knowledge entity: %w", knowledgeErr)
	}

	// TODO: use regular user service update flow here instead

	queryLinked := s.db.Client(ctx).User.Query().
		Where(user.KnowledgeEntityID(keId))
	linked, linkedErr := queryLinked.Only(ctx)
	if linkedErr != nil && !ent.IsNotFound(linkedErr) {
		return nil, fmt.Errorf("query linked user: %w", linkedErr)
	}

	var emailUser *ent.User
	if linked == nil || linked.Email != attrs.Email {
		var emailErr error
		emailUser, emailErr = s.db.Client(ctx).User.Query().
			Where(user.Email(attrs.Email)).
			Only(ctx)
		if emailErr != nil && !ent.IsNotFound(emailErr) {
			return nil, fmt.Errorf("query email user: %w", emailErr)
		}
	}

	var userID uuid.UUID
	if linked != nil {
		userID = linked.ID
	}
	if emailUser != nil {
		if userID != uuid.Nil && userID != emailUser.ID {
			return nil, fmt.Errorf("knowledge entity %s is linked to user %s but email %q belongs to user %s",
				keId,
				userID,
				attrs.Email,
				emailUser.ID,
			)
		}
		userID = emailUser.ID
	}

	var mutator ent.EntityMutator[*ent.User, *ent.UserMutation]
	if userID == uuid.Nil {
		mutator = s.db.Client(ctx).User.Create()
	} else {
		mutator = s.db.Client(ctx).User.UpdateOneID(userID)
	}

	m := mutator.Mutation()
	m.SetKnowledgeEntityID(keId)
	m.SetName(attrs.Name)
	m.SetEmail(attrs.Email)
	m.SetChatID(attrs.ChatId)
	m.SetTimezone(attrs.Timezone)

	saved, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save user: %w", saveErr)
	}
	projIds := map[string][]uuid.UUID{
		"user": {saved.ID},
	}

	return projIds, nil
}
