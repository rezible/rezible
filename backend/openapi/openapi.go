package openapi

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type (
	API         = huma.API
	Context     = huma.Context
	Config      = huma.Config
	ErrorModel  = huma.ErrorModel
	ErrorDetail = huma.ErrorDetail
	StatusError = huma.StatusError
	Operation   = huma.Operation
	Adapter     = huma.Adapter
	Middleware  = func(Context, func(Context))
)

func Register[I any, O any](api API, op Operation, handler func(context.Context, *I) (*O, error)) {
	huma.Register(api, op, handler)
}
