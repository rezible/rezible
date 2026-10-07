package apiv1

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test/mocks"
)

type installedWithoutHealthCheck struct{}

func (installedWithoutHealthCheck) Integration() *ent.Integration {
	return &ent.Integration{Name: "without_health_check"}
}

func (installedWithoutHealthCheck) Config() rez.IntegrationInstallationConfig {
	return nil
}

func (installedWithoutHealthCheck) Capabilities() []string {
	return nil
}

type installedWithHealthCheck struct {
	installedWithoutHealthCheck
	healthErr error
}

func (i installedWithHealthCheck) CheckHealth(context.Context) error {
	return i.healthErr
}

func TestCheckIntegrationHealth(t *testing.T) {
	for _, tc := range []struct {
		name      string
		installed rez.InstalledIntegration
		expected  oapi.IntegrationHealthCheck
	}{
		{
			name:      "a passing check is ok",
			installed: installedWithHealthCheck{},
			expected:  oapi.IntegrationHealthCheck{Ok: true},
		},
		{
			name:      "a failed check is ok false with the error text",
			installed: installedWithHealthCheck{healthErr: errors.New("logs: no logs data source configured")},
			expected:  oapi.IntegrationHealthCheck{Ok: false, Error: "logs: no logs data source configured"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			id := uuid.New()
			service := mocks.NewMockIntegrationService(t)
			service.EXPECT().GetInstalledIntegration(ctx, id).Return(tc.installed, nil)
			handler := newIntegrationsHandler(service)

			response, checkErr := handler.CheckIntegrationHealth(ctx, &oapi.CheckIntegrationHealthRequest{Id: id})

			require.NoError(t, checkErr)
			require.Equal(t, tc.expected, response.Body.Data)
		})
	}

	t.Run("an integration without the capability is not supported", func(t *testing.T) {
		ctx := t.Context()
		id := uuid.New()
		service := mocks.NewMockIntegrationService(t)
		service.EXPECT().GetInstalledIntegration(ctx, id).Return(installedWithoutHealthCheck{}, nil)
		handler := newIntegrationsHandler(service)

		response, checkErr := handler.CheckIntegrationHealth(ctx, &oapi.CheckIntegrationHealthRequest{Id: id})

		require.Nil(t, response)
		var statusErr huma.StatusError
		require.ErrorAs(t, checkErr, &statusErr)
		require.Equal(t, http.StatusUnprocessableEntity, statusErr.GetStatus())
	})
}
