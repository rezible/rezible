package openapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/rezible/rezible"
)

type (
	API         = huma.API
	Context     = huma.Context
	ErrorModel  = huma.ErrorModel
	StatusError = huma.StatusError
	Adapter     = huma.Adapter
	Middleware  = func(Context, func(Context))
)

type Handler interface {
	// GetMiddleware() []Middleware

	OncallMetricsHandler

	SystemComponentsHandler
	SystemAnalysisHandler

	EnvironmentsHandler
	FunctionalitiesHandler

	IncidentsHandler
	IncidentMilestonesHandler
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
	PlaybooksHandler
	MeetingsHandler

	AlertsHandler
	OncallHandler
	OncallEventsHandler

	UsersHandler
	AuthSessionsHandler
	TeamsHandler
	SubscriptionsHandler

	IntegrationsHandler
}
type operations struct{ Handler }

func MakeConfig() huma.Config {
	cfg := huma.DefaultConfig("Rezible API", "0.0.1")
	cfg.DocsPath = ""
	cfg.Servers = []*huma.Server{
		{URL: rez.BackendUrl},
	}
	cfg.Info.Description = "Rezible API Specification"

	cfg.Security = DefaultSecurity
	cfg.Components.SecuritySchemes = DefaultSecuritySchemes

	return cfg
}

func RegisterRoutes(api huma.API, handler Handler) {
	huma.AutoRegister(api, operations{handler})
}

func MakeApi(s Handler, prefix string, mw ...Middleware) huma.API {
	cfg := MakeConfig()
	/*
		cfg.Transformers = append([]huma.Transformer{
			interceptErrors(s)},
			cfg.Transformers...,
		)
	*/

	adapter := humago.NewAdapter(http.NewServeMux(), prefix)
	api := huma.NewAPI(cfg, adapter)
	api.UseMiddleware(mw...)
	RegisterRoutes(api, s)

	return api
}
