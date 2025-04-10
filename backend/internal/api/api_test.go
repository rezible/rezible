package api

import (
	"fmt"
	"testing"

	oapi "github.com/rezible/rezible/openapi"
)

func TestApiHandler(t *testing.T) {
	h := Handler{}
	api := oapi.MakeTestAPI(t, oapi.MakeConfig(), h)

	for p, r := range api.OpenAPI().Paths {
		fmt.Printf("p:%s, r: %+v\n", p, r.Get)
		break
	}
}
