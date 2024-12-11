package openapi

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/twohundreds/rezible"
)

type (
	Context     = huma.Context
	ErrorModel  = huma.ErrorModel
	StatusError = huma.StatusError
	Middleware  func(ctx Context, next func(Context))
)

type Handler interface {
	// GetMiddleware() []Middleware

	EnvironmentsHandler
	ServicesHandler
	FunctionalitiesHandler

	IncidentsHandler
	IncidentEventsHandler
	IncidentFieldsHandler
	IncidentRolesHandler
	IncidentTagsHandler
	IncidentTypesHandler
	IncidentSeveritiesHandler
	IncidentDebriefsHandler

	DocumentsHandler
	RetrospectivesHandler
	TasksHandler
	MeetingsHandler

	OncallHandler

	// LaddersHandler

	UsersHandler
	UserSessionsHandler
	TeamsHandler
	SubscriptionsHandler

	IntegrationsHandler
}
type operations struct{ Handler }

func DefaultConfig() huma.Config {
	cfg := huma.DefaultConfig("Rezible API", "0.0.1")
	cfg.DocsPath = ""
	cfg.Servers = []*huma.Server{
		{URL: fmt.Sprintf("%s/api/v1", rez.BackendUrl)},
	}
	cfg.Info.Description = "Rezible API Specification"
	return cfg
}

func RegisterRoutes(api huma.API, handler Handler) {
	huma.AutoRegister(api, operations{handler})
}

func MakeDefaultApi(s Handler) huma.API {
	cfg := DefaultConfig()
	/*
		cfg.Transformers = append([]huma.Transformer{
			interceptErrors(s)},
			cfg.Transformers...,
		)
	*/

	adapter := humago.NewAdapter(http.NewServeMux(), "")
	api := huma.NewAPI(cfg, adapter)
	RegisterRoutes(api, s)

	return api
}
