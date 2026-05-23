package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/integrations/projections"
)

func handleUserEventProjection(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	slog.Debug("user event projection handler", "kind", event.SubjectKind)
	if projections.SubjectKindUser.Matches(event) {
		projEvent, eventErr := projections.DecodeUserEvent(event)
		if eventErr != nil || projEvent == nil {
			return fmt.Errorf("invalid event: %w", eventErr)
		}
		return (&userEventProjectionHandler{client: client, projEvent: projEvent}).handle(ctx)
	}

	return nil
}

type userEventProjectionHandler struct {
	client    *ent.Client
	projEvent *projections.UserEvent
}

func (h *userEventProjectionHandler) handle(ctx context.Context) error {
	attrs := h.projEvent.Attributes

	create := h.client.User.Create().
		SetName(attrs.Name).
		SetEmail(attrs.Email).
		SetChatID(attrs.ChatId).
		SetTimezone(attrs.Timezone)
	upsert := create.OnConflictColumns(user.FieldTenantID, user.FieldEmail).
		UpdateNewValues()

	if saveErr := upsert.Exec(ctx); saveErr != nil {
		return fmt.Errorf("failed to update user: %w", saveErr)
	}

	return nil
}
