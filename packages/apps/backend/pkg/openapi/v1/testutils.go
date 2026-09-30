package v1

import (
	"testing"

	"github.com/rezible/rezible/pkg/openapi"
)

func NewTestAPI(t *testing.T, api API) *openapi.TestAPI {
	t.Helper()
	return openapi.NewTestAPI(t, api, VersionPrefix)
}
