package openapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
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

	EventsHandler
	EventAnnotationsHandler
	
	AlertsHandler

	OncallRostersHandler
	OncallShiftsHandler

	UsersHandler
	AuthSessionsHandler
	TeamsHandler

	IntegrationsHandler
}
type operations struct{ Handler }

func MakeConfig() huma.Config {
	cfg := huma.DefaultConfig("Rezible API", "0.0.1")
	cfg.DocsPath = ""
	cfg.Servers = []*huma.Server{
		//{URL: rez.BackendUrl},
	}
	cfg.Info.Description = "Rezible API Specification"

	cfg.Security = DefaultSecurity
	cfg.Components.SecuritySchemes = DefaultSecuritySchemes

	return cfg
}

func RegisterRoutes(api huma.API, handler Handler) {
	huma.AutoRegister(api, operations{Handler: handler})
}

func MakeApi(s Handler, prefix string, mw ...Middleware) huma.API {
	cfg := MakeConfig()

	//tranformers := []huma.Transformer{
	//	interceptErrors(s),
	//}
	//cfg.Transformers = append(cfg.Transformers, tranformers...)

	adapter := humago.NewAdapter(http.NewServeMux(), prefix)
	api := huma.NewAPI(cfg, adapter)
	api.UseMiddleware(mw...)
	RegisterRoutes(api, s)

	return api
}
