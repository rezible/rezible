package v1

import (
	"net/http"

	"github.com/rezible/rezible/openapi"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"gopkg.in/yaml.v3"
)

const versionPrefix = "/v1"

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

func MakeApi(h Handler, prefix string, mw ...openapi.Middleware) openapi.API {
	cfg := MakeConfig()

	//tranformers := []huma.Transformer{
	//	interceptErrors(s),
	//}
	//cfg.Transformers = append(cfg.Transformers, tranformers...)

	adapter := humago.NewAdapter(http.NewServeMux(), prefix+versionPrefix)
	api := huma.NewAPI(cfg, adapter)
	api.UseMiddleware(mw...)
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

	cfg.Security = openapi.DefaultSecurity
	cfg.Components.SecuritySchemes = openapi.DefaultSecuritySchemes

	return cfg
}

func GetYamlSpec() (string, error) {
	api := MakeApi(operations{}, versionPrefix)
	spec, specErr := yaml.Marshal(api.OpenAPI())
	if specErr != nil {
		return "", specErr
	}
	return string(spec), nil
}
