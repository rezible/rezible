package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/integrations/projections"
)

func userEventProjectionHandler(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	slog.Debug("user event projection handler", "kind", event.Kind.String())
	if event.Kind != ne.KindUserObserved {
		return nil
	}
	decoded, validationErr := projections.DecodeEvent[projections.UserObservedAttributes](event)
	if validationErr != nil || decoded == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}

	attrs := decoded.Attributes
	upsert := client.Debug().User.Create().
		SetName(attrs.Name).
		SetEmail(attrs.Email).
		SetChatID(attrs.ChatId).
		SetTimezone(attrs.Timezone).
		OnConflictColumns(user.FieldTenantID, user.FieldEmail).
		Update(func(u *ent.UserUpsert) {
			u.UpdateTimezone()
			u.UpdateChatID()
		})
	if saveErr := upsert.Exec(ctx); saveErr != nil {
		return fmt.Errorf("failed to update user: %w", saveErr)
	}

	return nil
}
