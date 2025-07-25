package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"reflect"
)

const (
	DefaultListLimit = 10
)

type Expandable[Attrs any] struct {
	Id         uuid.UUID `json:"id"`
	Attributes *Attrs    `json:"attributes,omitempty"`
}

// Requests
type (
	EmptyRequest struct{}
	ListRequest  struct {
		Limit           int    `query:"limit" maximum:"50" minimum:"1" default:"10" required:"false" nullable:"false"`
		Offset          int    `query:"offset" minimum:"0" default:"0" required:"false" nullable:"false"`
		Search          string `query:"search" required:"false" nullable:"false"`
		IncludeArchived bool   `query:"archived" required:"false" nullable:"false" default:"false"`
		// Sort   string  `query:"sort" enum:"asc,desc" default:"asc" required:"false" nullable:"false"`
	}
	ListIdRequest struct {
		Id uuid.UUID `path:"id"`
		ListRequest
	}
	RequestWithBodyAttributes[T any] struct {
		Body struct {
			Attributes T `json:"attributes"`
		}
	}
	PostIdRequest struct {
		Id uuid.UUID `path:"id"`
	}
	CreateIdRequest[T any] struct {
		Id uuid.UUID `path:"id"`
		RequestWithBodyAttributes[T]
	}
	GetIdRequest struct {
		Id uuid.UUID `path:"id"`
	}
	GetFlexibleIdRequest struct {
		PathId string `path:"id"`
		Id     FlexibleId
	}
	UpdateIdRequest[T any] struct {
		Id uuid.UUID `path:"id"`
		RequestWithBodyAttributes[T]
	}
	DeleteIdRequest struct {
		Id uuid.UUID `path:"id"`
	}
	ArchiveIdRequest struct {
		Id uuid.UUID `path:"id"`
	}
)

func (l ListRequest) ListParams() rez.ListParams {
	return rez.ListParams{
		Search:          l.Search,
		Offset:          l.Offset,
		Limit:           l.Limit,
		IncludeArchived: l.IncludeArchived,
		Count:           true,
	}
}

// Responses
type (
	EmptyResponse       struct{}
	ItemResponse[T any] struct {
		Body struct {
			Data T `json:"data"`
		}
	}
	ListResponse[T any] struct {
		Body struct {
			Data []T `json:"data"`
		}
	}
	PaginatedResponse[T any] struct {
		Body struct {
			Data       []T                `json:"data"`
			Pagination ResponsePagination `json:"pagination"`
		}
	}
	ResponsePagination struct {
		Next     *string `json:"next,omitempty"`
		Previous *string `json:"previous,omitempty"`
		Total    int     `json:"total"`
	}
)

// FlexibleId is a field which can be either a uuid or a slug
type FlexibleId struct {
	IsUUID bool
	UUID   uuid.UUID
	IsSlug bool
	Slug   string
}

func GetEntPredicate[P any](id FlexibleId, uuidPred func(uuid.UUID) P, slugPred func(string) P) P {
	if id.IsUUID {
		return uuidPred(id.UUID)
	} else {
		return slugPred(id.Slug)
	}
}

func resolveFlexibleId(idParam string) (*FlexibleId, error) {
	uid, parseErr := uuid.Parse(idParam)
	if parseErr == nil {
		return &FlexibleId{IsUUID: true, UUID: uid}, nil
	} else if len(idParam) > 1 { // TODO: min slug length
		return &FlexibleId{IsSlug: true, Slug: idParam}, nil
	} else {
		return nil, errors.New("invalid id")
	}
}

func (i *GetFlexibleIdRequest) Resolve(ctx huma.Context) []error {
	idParam := ctx.Param("id")
	id, parseErr := resolveFlexibleId(idParam)
	if parseErr != nil {
		return []error{parseErr}
	}
	i.Id = *id
	return nil
}

// OmittableNullable is a field which can be omitted from the input,
// set to `null`, or set to a value. Each state is tracked and can
// be checked for in handling code.
type OmittableNullable[T any] struct {
	Sent  bool
	Null  bool
	Value T
}

func (o *OmittableNullable[T]) NillableValue() *T {
	if !o.Sent {
		return nil
	}
	if o.Null {
		var emptyValue T
		return &emptyValue
	}
	return &o.Value
}

func (o *OmittableNullable[T]) UnmarshalJSON(b []byte) error {
	if len(b) > 0 {
		o.Sent = true
		if bytes.Equal(b, []byte("null")) {
			o.Null = true
			return nil
		}
		return json.Unmarshal(b, &o.Value)
	}
	return nil
}

func (o OmittableNullable[T]) Schema(r huma.Registry) *huma.Schema {
	s := r.Schema(reflect.TypeOf(o.Value), true, "")
	s.Extensions = map[string]interface{}{
		"nullable": true,
	}
	return s
}

func WrapContext(ctx Context, sub context.Context) Context {
	return huma.WithContext(ctx, sub)
}

func WrapContextWithValue(ctx Context, key any, value any) Context {
	return huma.WithValue(ctx, key, value)
}
