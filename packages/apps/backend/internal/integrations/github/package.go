package github

import (
	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(func(i do.Injector) (*Integration, error) {
		return MakeIntegration(
			do.MustInvoke[rez.ConfigLoader](i),
			do.MustInvoke[rez.ProviderEventService](i),
		)
	}),
)
