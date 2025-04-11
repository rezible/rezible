package openapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	AnalyticsHandler

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
	MeetingsHandler

	OncallHandler
	OncallEventsHandler

	UsersHandler
	AuthSessionsHandler
	TeamsHandler
	SubscriptionsHandler

	IntegrationsHandler
}
type operations struct{ Handler }

const SessionCookieName = "rez_session"

func MakeConfig() huma.Config {
	cfg := huma.DefaultConfig("Rezible API", "0.0.1")
	cfg.DocsPath = ""
	cfg.Servers = []*huma.Server{
		{URL: fmt.Sprintf("%s/api/v1", rez.BackendUrl)},
	}
	cfg.Info.Description = "Rezible API Specification"

	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"session-cookie": {
			Type: "apiKey",
			In:   "cookie",
			Name: SessionCookieName,
		},
		"api-token": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
		"session-token": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	cfg.Security = []map[string][]string{
		{"session-cookie": {}},
		{"api-token": {}},
	}

	return cfg
}

func RegisterRoutes(api huma.API, handler Handler) {
	huma.AutoRegister(api, operations{handler})
}

func MakeApi(s Handler, mw ...Middleware) huma.API {
	cfg := MakeConfig()
	/*
		cfg.Transformers = append([]huma.Transformer{
			interceptErrors(s)},
			cfg.Transformers...,
		)
	*/

	adapter := humago.NewAdapter(http.NewServeMux(), "")
	api := huma.NewAPI(cfg, adapter)
	api.UseMiddleware(mw...)
	RegisterRoutes(api, s)

	return api
}

func Unwrap(c Context) (*http.Request, http.ResponseWriter) {
	return humago.Unwrap(c)
}

func WithContext(c Context, ctx context.Context) Context {
	return huma.WithContext(c, ctx)
}

func WriteAuthError(w http.ResponseWriter, authErr error) error {
	var resp StatusError
	if errors.Is(authErr, rez.ErrNoAuthSession) {
		resp = ErrorUnauthorized("no_session")
	} else if errors.Is(authErr, rez.ErrAuthSessionExpired) {
		resp = ErrorUnauthorized("session_expired")
	} else if errors.Is(authErr, rez.ErrAuthSessionUserMissing) {
		resp = ErrorUnauthorized("missing_user")
	} else {
		resp = ErrorUnauthorized("unknown")
	}
	respBody, jsonErr := json.Marshal(resp)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(resp.GetStatus())
	_, writeErr := w.Write(respBody)
	return writeErr
}
