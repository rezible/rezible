package v1

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/rezible/rezible/openapi"
)

const VersionPrefix = "/v1"

type Handler interface {
	// GetMiddleware() []Middleware

	UserSessionsHandler
	OrganizationsHandler
	UsersHandler
	TeamsHandler
	IntegrationsHandler
	AgentRunsHandler

	OncallMetricsHandler

	SystemTopologyHandler
	SystemAnalysisHandler

	IncidentsHandler
	IncidentMetadataHandler
	IncidentMilestonesHandler
	IncidentTimelineHandler
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
}
type operations struct{ Handler }

func makeConfig() openapi.Config {
	cfg := huma.DefaultConfig("Rezible API", "0.0.1")
	cfg.DocsPath = ""
	cfg.OpenAPIPath = "/openapi"
	cfg.Servers = []*huma.Server{
		//{
		//	URL:         "https://app.dev.rezible.com/api/v1",
		//	Description: "Local Development",
		//},
	}
	cfg.Info.Description = "Rezible API Specification"
	cfg.Security = DefaultSecurityMethods
	cfg.Components.SecuritySchemes = MethodSecuritySchemes()

	return cfg
}

func MakeApi(h Handler, middlewares ...openapi.Middleware) openapi.API {
	api := humago.NewWithPrefix(http.NewServeMux(), VersionPrefix, makeConfig())
	api.UseMiddleware(middlewares...)
	huma.AutoRegister(api, operations{Handler: h})
	return api
}

func MakeOpenApiSpec() *huma.OpenAPI {
	return makeUnhandledApi().OpenAPI()
}

func makeUnhandledApi() openapi.API {
	return MakeApi(operations{})
}
