package apiv1

import (
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/require"

	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

func TestIncidentRoleMaxAssigneesMustBePositive(t *testing.T) {
	_, api := humatest.New(t)
	handler := newIncidentMetadataHandler(nil, nil)
	huma.Register(api, oapi.CreateIncidentRole, handler.CreateIncidentRole)
	huma.Register(api, oapi.UpdateIncidentRole, handler.UpdateIncidentRole)

	created := api.Post("/incident_metadata/roles", map[string]any{
		"attributes": map[string]any{"name": "Commander", "maxAssignees": 0},
	})
	require.Equal(t, http.StatusUnprocessableEntity, created.Code, created.Body.String())

	updated := api.Patch("/incident_metadata/roles/7d6f3c1e-8a52-4a8f-9d43-1f0e0b6f2a11", map[string]any{
		"attributes": map[string]any{"maxAssignees": -1},
	})
	require.Equal(t, http.StatusUnprocessableEntity, updated.Code, updated.Body.String())
}
