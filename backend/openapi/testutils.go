///go:build testing

package openapi

import (
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

func MakeTestAPI(t *testing.T, cfg huma.Config, s any) (http.Handler, API) {
	h, api := humatest.New(t, cfg)
	huma.AutoRegister(api, s)
	return h, api
}
