package v1

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/openapi"
)

const VersionPrefix = "/v1"

type Handler interface {
	// GetMiddleware() []Middleware

	AuthSessionsHandler
	OrganizationsHandler
	UsersHandler
	TeamsHandler
	IntegrationsHandler

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
}
type operations struct{ Handler }

func MakeApi(h Handler, auth rez.AuthService) openapi.API {
	cfg := MakeConfig()

	//tranformers := []huma.Transformer{
	//	interceptErrors(s),
	//}
	//cfg.Transformers = append(cfg.Transformers, tranformers...)

	adapter := humago.NewAdapter(http.NewServeMux(), VersionPrefix)
	api := huma.NewAPI(cfg, adapter)
	api.UseMiddleware(MakeSecurityMiddleware(api, auth))
	huma.AutoRegister(api, operations{Handler: h})

	return api
}

func MakeConfig() openapi.Config {
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
	cfg.Security = ApiSecurityMethods
	cfg.Components.SecuritySchemes = GetDefaultSecuritySchemes()

	return cfg
}

func GetSpec(jsonFmt bool) ([]byte, error) {
	spec := MakeApi(operations{}, nil).OpenAPI()
	if jsonFmt {
		return spec.MarshalJSON()
	}
	return spec.YAML()
}
