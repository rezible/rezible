package v1

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/rezible/rezible/pkg/openapi"
)

const VersionPrefix = "/v1"

type (
	Handler interface {
		SecurityProvider

		ActivityHandler
		UserSessionsHandler
		OrganizationsHandler
		UsersHandler
		TeamsHandler
		IntegrationsHandler
		AiHandler

		OncallMetricsHandler

		KnowledgeGraphHandler
		SystemAnalysisHandler
		SituationsHandler
		InvestigationsHandler

		IncidentsHandler
		ReviewsHandler
		IncidentMetadataHandler
		IncidentMilestonesHandler
		IncidentDebriefsHandler

		DocumentsHandler
		RetrospectivesHandler
		TasksHandler
		PlaybooksHandler
		MeetingsHandler

		EventsHandler
		EventAnnotationsHandler

		AlertsHandler
		DiscussionHandler

		OncallRostersHandler
		OncallShiftsHandler
	}

	operations struct{ Handler }

	API openapi.API
)

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
	cfg.Info.Description = "Rezible API"
	cfg.Security = DefaultOperationSecurityMethodOptions
	cfg.Components.SecuritySchemes = MethodSecuritySchemes()

	return cfg
}

func MakeApi(h Handler, middlewares ...openapi.Middleware) API {
	o := operations{Handler: h}

	api := humago.NewWithPrefix(http.NewServeMux(), VersionPrefix, makeConfig())
	api.UseMiddleware(makeRequestMethodSecurityMiddleware(api, o, humago.Unwrap))
	api.UseMiddleware(middlewares...)

	autoRegisterEnumAliases(api, o)
	huma.AutoRegister(api, o)

	return api
}

func makeUnhandledApi() API {
	return MakeApi(operations{})
}

type SpecFormat string

const (
	SpecFormatJSON SpecFormat = "json"
	SpecFormatYAML SpecFormat = "yaml"
)

func encodeSpec(f SpecFormat, spec *openapi.OpenAPI) ([]byte, error) {
	switch f {
	case SpecFormatJSON:
		return spec.MarshalJSON()
	case SpecFormatYAML:
		return spec.YAML()
	default:
		return nil, fmt.Errorf("invalid spec format")
	}
}

func GetEncodedSpec(f SpecFormat) ([]byte, error) {
	return encodeSpec(f, makeUnhandledApi().OpenAPI())
}
