package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type ProviderResourceRef struct {
	Provider          string `json:"provider"`
	ProviderNamespace string `json:"providerNamespace"`
	ResourceRef       string `json:"resourceRef"`
}

func ProviderResourceRefFromRez(ref rez.ProviderResourceRef) ProviderResourceRef {
	return ProviderResourceRef{
		Provider:          ref.Provider,
		ProviderNamespace: ref.ProviderNamespace,
		ResourceRef:       ref.ResourceRef,
	}
}

type Expandable[Attrs any] struct {
	Id         uuid.UUID `json:"id"`
	Attributes *Attrs    `json:"attributes,omitempty"`
}

// Requests
type (
	EmptyRequest      struct{}
	PaginationRequest struct {
		Page     int `query:"page" minimum:"1" default:"1" required:"false" nullable:"false"`
		PageSize int `query:"pageSize" maximum:"50" minimum:"1" default:"25" required:"false" nullable:"false"`
	}
	RequestWithBodyAttributes[T any] struct {
		Body struct {
			Attributes T `json:"attributes"`
		}
	}

	IdRequest struct {
		Id uuid.UUID `path:"id"`
	}
	IdRequestWithBody[T any] struct {
		Id uuid.UUID `path:"id"`
		RequestWithBodyAttributes[T]
	}
	PaginatedIdRequest struct {
		Id uuid.UUID `path:"id"`
		PaginationRequest
	}

	FlexibleIdRequest struct {
		PathId string `path:"id"`
		Id     FlexibleId
	}

	EmptyNameRequest struct {
		Name string `path:"name"`
	}
	NameRequest[A any] struct {
		EmptyNameRequest
		RequestWithBodyAttributes[A]
	}
)

func (p PaginationRequest) ListParams() ent.ListParams {
	return ent.ListParams{
		Page:     p.Page,
		PageSize: p.PageSize,
	}
}

// Responses
type (
	EmptyResponse     struct{}
	SetCookieResponse struct {
		SetCookie []http.Cookie `header:"Set-Cookie"`
	}
	ItemResponse[T any] struct {
		Body struct {
			Data T `json:"data"`
		}
	}
	PaginatedResponse[T any] struct {
		Body PaginatedResponseBody[T]
	}
	PaginatedResponseBody[T any] struct {
		Data       []T        `json:"data" nullable:"false"`
		Pagination Pagination `json:"pagination"`
	}
	Pagination struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
		Total    int `json:"total"`
	}
	CollectionResponse[T any] struct {
		Body CollectionResponseBody[T]
	}
	CollectionResponseBody[T any] struct {
		Data []T `json:"data" nullable:"false"`
	}
)

func ConvertSlice[D any, O any](data []D, fn func(D) O) []O {
	res := make([]O, len(data))
	for i, d := range data {
		res[i] = fn(d)
	}
	return res
}

func ConvertResultPagination[T any](r *ent.ListResult[T]) Pagination {
	return Pagination{
		Page:     r.Page,
		PageSize: r.PageSize,
		Total:    r.Total,
	}
}

func ConvertPaginatedResultBody[T any, R any](result *ent.ListResult[R], fn func(*R) T) PaginatedResponseBody[T] {
	return PaginatedResponseBody[T]{
		Data:       ConvertSlice(result.Data, fn),
		Pagination: ConvertResultPagination(result),
	}
}

func MaybeConvertSlice[D any, O any](data []D, fn func(D) (*O, error)) ([]O, error) {
	res := make([]O, len(data))
	for i, d := range data {
		rd, fnErr := fn(d)
		if fnErr != nil {
			return nil, fnErr
		}
		res[i] = *rd
	}
	return res, nil
}

func MaybeConvertPaginatedResultBody[T any, R any](result *ent.ListResult[R], fn func(*R) (*T, error)) (*PaginatedResponseBody[T], error) {
	data, dataErr := MaybeConvertSlice(result.Data, fn)
	if dataErr != nil {
		return nil, dataErr
	}
	return &PaginatedResponseBody[T]{Data: data, Pagination: ConvertResultPagination(result)}, nil
}

type CalendarDate string

func (r CalendarDate) Parse() (time.Time, error) {
	return time.Parse("2006-01-02", string(r))
}

type CalendarDateTime string

func (r CalendarDateTime) Parse() (time.Time, error) {
	return time.Parse("2006-01-02", string(r))
}

func GetCalendarDateWindow(from, to CalendarDate) (time.Time, time.Time, error) {
	parsedFrom, fromErr := from.Parse()
	parsedTo, toErr := to.Parse()
	return parsedFrom, parsedTo, errors.Join(fromErr, toErr)
}

// FlexibleId is a field which can be either a UUID or a slug
type FlexibleId struct {
	IsUUID bool
	UUID   uuid.UUID
	IsSlug bool
	Slug   string
}

func (i *FlexibleIdRequest) Resolve(ctx huma.Context) []error {
	idParam := ctx.Param("id")
	uid, parseErr := uuid.Parse(idParam)
	if parseErr == nil {
		i.Id = FlexibleId{IsUUID: true, UUID: uid}
	} else if len(idParam) > 1 {
		// TODO: min slug length
		i.Id = FlexibleId{IsSlug: true, Slug: idParam}
	} else {
		return []error{fmt.Errorf("invalid id param")}
	}
	return nil
}

type OptionalParam[T any] struct {
	Value T
	IsSet bool
}

func (o OptionalParam[T]) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeOf(o.Value))
}

func (o *OptionalParam[T]) Receiver() reflect.Value {
	return reflect.ValueOf(o).Elem().Field(0)
}

func (o *OptionalParam[T]) OnParamSet(isSet bool, parsed any) {
	o.IsSet = isSet
}

// OmittableNullable is a field which can be omitted from the input,
// set to `null`, or set to a value. Each state is tracked and can
// be checked for in handling code.
type OmittableNullable[T any] struct {
	Sent  bool
	Null  bool
	Value T
}

func (o OmittableNullable[T]) NillableValue() *T {
	if !o.Sent {
		return nil
	}
	if o.Null {
		var emptyValue T
		return &emptyValue
	}
	return &o.Value
}

func (o OmittableNullable[T]) UnmarshalJSON(b []byte) error {
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
