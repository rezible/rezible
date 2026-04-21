package openapi

import (
	"github.com/danielgtaylor/huma/v2"
)

//go:generate go run -mod=mod generate_specs.go

type (
	API        = huma.API
	Context    = huma.Context
	Config     = huma.Config
	Operation  = huma.Operation
	Adapter    = huma.Adapter
	Middleware = func(Context, func(Context))
)

func init() {
	huma.DefaultArrayNullable = false
}
