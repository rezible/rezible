package watermill

import (
	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(func(i do.Injector) (rez.MessageService, error) {
		return NewMessageService(do.MustInvoke[rez.TelemetryService](i))
	}),
)
