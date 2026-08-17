package projections

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
)

type Event[T any] struct {
	Event      *ent.NormalizedEvent
	Attributes T
}

type EventDisplay struct {
	Title       string
	Description string
}

func GetEventDisplay(ev *ent.NormalizedEvent) (*EventDisplay, error) {
	if ev == nil {
		return nil, fmt.Errorf("normalized event is nil")
	}
	// TODO: actually convert
	disp := &EventDisplay{
		Title:       "TODO",
		Description: "",
	}
	return disp, nil
}

func EncodeAttributes[A any](attrs A) ([]byte, error) {
	if validationErr := validateAttributes(attrs); validationErr != nil {
		return nil, fmt.Errorf("validate attributes: %w", validationErr)
	}
	attrBytes, marshalErr := json.Marshal(attrs)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal attributes: %w", marshalErr)
	}
	return attrBytes, nil
}

func DecodeSubjectAttributes[A any](ev *ent.NormalizedEvent) (*Event[A], error) {
	if ev == nil {
		return nil, fmt.Errorf("normalized event is nil")
	}
	var attrs A
	if decodeErr := json.Unmarshal(ev.Attributes, &attrs); decodeErr != nil {
		return nil, fmt.Errorf("decode attributes: %w", decodeErr)
	}
	if validationErr := validateAttributes(attrs); validationErr != nil {
		return nil, fmt.Errorf("validate attributes: %w", validationErr)
	}
	return &Event[A]{Event: ev, Attributes: attrs}, nil
}

var attributeValidator = newProjectionValidator()

var ErrRetryableProjection = errors.New("retryable projection failure")

func Retryable(err error) error {
	if err == nil {
		return ErrRetryableProjection
	}
	return fmt.Errorf("%w: %w", ErrRetryableProjection, err)
}

func IsRetryable(err error) bool {
	return errors.Is(err, ErrRetryableProjection)
}

func newProjectionValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, found := strings.Cut(field.Tag.Get("json"), ",")
		if found && name != "" && name != "-" {
			return name
		}
		return field.Name
	})
	return validate
}

func validateAttributes[A any](attrs A) error {
	if validationErr := attributeValidator.Struct(attrs); validationErr != nil {
		if errs, ok := errors.AsType[validator.ValidationErrors](validationErr); ok && len(errs) > 0 {
			msgs := make([]string, len(errs))
			for i, verr := range errs {
				msgs[i] = fmt.Sprintf("[%s:%s]", verr.Field(), verr.Error())
			}
			return fmt.Errorf("field: %s", strings.Join(msgs, " "))
		}
	}
	return nil
}

type RelatedEntityRef struct {
	ExternalRef string   `json:"external_ref" validate:"required"`
	Kind        kne.Kind `json:"kind" validate:"required"`
	Subkind     string   `json:"subkind" validate:"required"`
	DisplayName string   `json:"display_name" validate:"required"`
}

func SortRelatedEntityRefs(refs []RelatedEntityRef) []RelatedEntityRef {
	sortedRefs := append([]RelatedEntityRef(nil), refs...)
	slices.SortStableFunc(sortedRefs, func(left, right RelatedEntityRef) int {
		if left.ExternalRef != right.ExternalRef {
			return strings.Compare(left.ExternalRef, right.ExternalRef)
		}
		if left.Kind != right.Kind {
			return strings.Compare(left.Kind.String(), right.Kind.String())
		}
		if left.Subkind != right.Subkind {
			return strings.Compare(left.Subkind, right.Subkind)
		}
		return strings.Compare(left.DisplayName, right.DisplayName)
	})
	return sortedRefs
}
