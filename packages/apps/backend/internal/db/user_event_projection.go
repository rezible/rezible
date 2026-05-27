package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/integrations/projections"
)

const (
	assertionUserProfileObserved = "user_profile_observed"
	knowledgeKindUser            = "user"
)

func HandleUserEventProjection(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	if projections.SubjectKindUser.Matches(event) {
		projEvent, eventErr := projections.DecodeUserEvent(event)
		if eventErr != nil || projEvent == nil {
			return fmt.Errorf("invalid event: %w", eventErr)
		}
		handler := &userEventProjectionHandler{
			client:    client,
			projEvent: projEvent,
			knowledge: newKnowledgeService(client),
		}
		return handler.handle(ctx)
	}

	return nil
}

type userEventProjectionHandler struct {
	client    *ent.Client
	projEvent *projections.UserEvent
	knowledge *KnowledgeService
}

func (h *userEventProjectionHandler) handle(ctx context.Context) error {
	attrs := h.projEvent.Attributes
	ev := h.projEvent.Event

	entityParams := ResolveKnowledgeEntityParams{
		Event:       ev,
		Kind:        knowledgeKindUser,
		DisplayName: attrs.Name,
		Aliases:     []EntityAliasRef{entityAliasRefForEvent(ev, "")},
	}
	savedKnowledge, knowledgeErr := h.knowledge.ResolveEntityWithAssertion(ctx, entityParams, assertionUserProfileObserved)
	if knowledgeErr != nil {
		return fmt.Errorf("resolve user knowledge entity: %w", knowledgeErr)
	}

	linked, linkedErr := h.client.User.Query().
		Where(user.KnowledgeEntityID(savedKnowledge.Entity.ID)).
		Only(ctx)
	if linkedErr != nil && !ent.IsNotFound(linkedErr) {
		return fmt.Errorf("query linked user: %w", linkedErr)
	}

	var emailUser *ent.User
	if linked == nil || linked.Email != attrs.Email {
		var emailErr error
		emailUser, emailErr = h.client.User.Query().
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
				savedKnowledge.Entity.ID,
				userID,
				attrs.Email,
				emailUser.ID,
			)
		}
		userID = emailUser.ID
	}

	var mut *ent.UserMutation
	if userID == uuid.Nil {
		mut = h.client.User.Create().Mutation()
	} else {
		mut = h.client.User.UpdateOneID(userID).Mutation()
	}
	mut.SetKnowledgeEntityID(savedKnowledge.Entity.ID)
	mut.SetName(attrs.Name)
	mut.SetEmail(attrs.Email)
	mut.SetChatID(attrs.ChatId)
	mut.SetTimezone(attrs.Timezone)

	if _, saveErr := h.client.Mutate(ctx, mut); saveErr != nil {
		return fmt.Errorf("save user: %w", saveErr)
	}

	return nil
}
