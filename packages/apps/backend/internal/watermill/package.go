package watermill

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

func ProvideMessageService(ctx context.Context, inj do.Injector) (rez.MessageService, error) {
	return NewMessageService(ctx)
}
