package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/integrations/projections"
)

const (
	assertionUserProfileObserved = "user_profile_observed"
	knowledgeKindUser            = "user"
)

func (s *UserService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) error {
	if projections.SubjectKindUser.Matches(event) {
		decoded, eventErr := projections.DecodeUserEvent(event)
		if eventErr != nil || decoded == nil {
			return fmt.Errorf("invalid event: %w", eventErr)
		}
		return s.handleUserEventProjection(ctx, decoded)
	}
	return nil
}

func (s *UserService) handleUserEventProjection(ctx context.Context, ue *projections.UserEvent) error {
	attrs := ue.Attributes
	entityParams := rez.ResolveKnowledgeEntityParams{
		Event:             ue.Event,
		EvidenceAssertion: assertionUserProfileObserved,
		Entity: &ent.KnowledgeEntity{
			Kind:        knowledgeKindUser,
			DisplayName: attrs.Name,
		},
		Aliases: []*ent.KnowledgeEntityAlias{
			{Provider: ue.Event.Provider, ProviderSubjectRef: ue.Event.ProviderSubjectRef},
		},
	}
	knoEnt, knowledgeErr := s.knowledge.ResolveEntity(ctx, entityParams)
	if knowledgeErr != nil {
		return fmt.Errorf("resolve user knowledge entity: %w", knowledgeErr)
	}

	// TODO: use regular user service update flow here instead

	linked, linkedErr := s.client.User.Query().
		Where(user.KnowledgeEntityID(knoEnt.ID)).
		Only(ctx)
	if linkedErr != nil && !ent.IsNotFound(linkedErr) {
		return fmt.Errorf("query linked user: %w", linkedErr)
	}

	var emailUser *ent.User
	if linked == nil || linked.Email != attrs.Email {
		var emailErr error
		emailUser, emailErr = s.client.User.Query().
			Where(user.Email(attrs.Email)).
			Only(ctx)
		if emailErr != nil && !ent.IsNotFound(emailErr) {
			return fmt.Errorf("query email user: %w", emailErr)
		}
	}

	var userID uuid.UUID
	if linked != nil {
		userID = linked.ID
	}
	if emailUser != nil {
		if userID != uuid.Nil && userID != emailUser.ID {
			return fmt.Errorf("knowledge entity %s is linked to user %s but email %q belongs to user %s",
				knoEnt.ID,
				userID,
				attrs.Email,
				emailUser.ID,
			)
		}
		userID = emailUser.ID
	}

	var mut *ent.UserMutation
	if userID == uuid.Nil {
		mut = s.client.User.Create().Mutation()
	} else {
		mut = s.client.User.UpdateOneID(userID).Mutation()
	}
	mut.SetKnowledgeEntityID(knoEnt.ID)
	mut.SetName(attrs.Name)
	mut.SetEmail(attrs.Email)
	mut.SetChatID(attrs.ChatId)
	mut.SetTimezone(attrs.Timezone)

	if _, saveErr := s.client.Mutate(ctx, mut); saveErr != nil {
		return fmt.Errorf("save user: %w", saveErr)
	}

	return nil
}
