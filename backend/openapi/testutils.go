///go:build testing

package openapi

import (
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

func MakeTestAPI(t *testing.T, cfg huma.Config, h Handler) huma.API {
	_, api := humatest.New(t, cfg)
	RegisterRoutes(api, h)
	return api
}
