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
	"github.com/rezible/rezible/ent"
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
	PostIdEmptyRequest struct {
		Id uuid.UUID `path:"id"`
	}
	PostIdRequest[T any] struct {
		Id uuid.UUID `path:"id"`
		RequestWithBodyAttributes[T]
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

	NameRequest struct {
		Name string `path:"name"`
	}
	NameRequestWithAttributes[A any] struct {
		NameRequest
		RequestWithBodyAttributes[A]
	}
)

func (l ListRequest) ListParams() ent.ListParams {
	return ent.ListParams{
		Search:          l.Search,
		Offset:          l.Offset,
		Limit:           l.Limit,
		IncludeArchived: l.IncludeArchived,
		Count:           true,
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

func GetCalendarDateTimeWindow(from, to CalendarDateTime) (time.Time, time.Time, error) {
	parsedFrom, fromErr := from.Parse()
	parsedTo, toErr := to.Parse()
	return parsedFrom, parsedTo, errors.Join(fromErr, toErr)
}

// FlexibleId is a field which can be either a uuid or a slug
type FlexibleId struct {
	IsUUID bool
	UUID   uuid.UUID
	IsSlug bool
	Slug   string
}

func (i *GetFlexibleIdRequest) Resolve(ctx huma.Context) []error {
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
