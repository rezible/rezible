package projections

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/rezible/rezible/ent"
)

type Event[T any] struct {
	Event      *ent.NormalizedEvent
	Attributes T
}

type EventProjectionHandlerFunc = func(context.Context, *ent.Client, *ent.NormalizedEvent) error

type EventProjectionHandler struct {
	Handler      EventProjectionHandlerFunc
	SubjectKinds []SubjectKind
}

var projectionFuncsMu sync.RWMutex
var projectionFuncs = make(map[string]EventProjectionHandler)

func RegisterHandler(name string, handler EventProjectionHandlerFunc, subjectKinds ...SubjectKind) {
	projectionFuncsMu.Lock()
	defer projectionFuncsMu.Unlock()
	projectionFuncs[name] = EventProjectionHandler{
		Handler:      handler,
		SubjectKinds: subjectKinds,
	}
}

func GetHandlersFor(ev *ent.NormalizedEvent) map[string]EventProjectionHandlerFunc {
	projectionFuncsMu.RLock()
	defer projectionFuncsMu.RUnlock()

	handlers := make(map[string]EventProjectionHandlerFunc)
	if ev == nil {
		return handlers
	}
	for name, registered := range projectionFuncs {
		if len(registered.SubjectKinds) == 0 {
			handlers[name] = registered.Handler
			continue
		}
		for _, subjectKind := range registered.SubjectKinds {
			if subjectKind.Matches(ev) {
				handlers[name] = registered.Handler
				break
			}
		}
	}
	return handlers
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

type EventAttributes map[string]any

func EncodeAttributes[A any](attrs A) (EventAttributes, error) {
	if validationErr := validateAttributes(attrs); validationErr != nil {
		return nil, fmt.Errorf("validate attributes: %w", validationErr)
	}
	// mapstructure encode is just a "reverse decode"
	var encodedAttrs EventAttributes
	cfg := &mapstructure.DecoderConfig{
		Result:      &encodedAttrs,
		TagName:     attributeFieldNameTag,
		ErrorUnused: true,
	}
	decoder, decoderErr := mapstructure.NewDecoder(cfg)
	if decoderErr != nil {
		return nil, fmt.Errorf("create decoder: %w", decoderErr)
	}
	if decodeErr := decoder.Decode(attrs); decodeErr != nil {
		return nil, fmt.Errorf("decode attributes: %w", decodeErr)
	}
	return encodedAttrs, nil
}

func DecodeSubjectAttributes[A any](ev *ent.NormalizedEvent) (*Event[A], error) {
	if ev == nil {
		return nil, fmt.Errorf("normalized event is nil")
	}
	eventAttrs := ev.Attributes
	if eventAttrs == nil {
		eventAttrs = EventAttributes{}
	}
	var attrs A
	cfg := &mapstructure.DecoderConfig{
		Result:      &attrs,
		TagName:     attributeFieldNameTag,
		ErrorUnused: true,
	}
	decoder, decoderErr := mapstructure.NewDecoder(cfg)
	if decoderErr != nil {
		return nil, fmt.Errorf("create decoder: %w", decoderErr)
	}
	if decodeErr := decoder.Decode(eventAttrs); decodeErr != nil {
		return nil, fmt.Errorf("decode attributes: %w", decodeErr)
	}
	if validationErr := validateAttributes(attrs); validationErr != nil {
		return nil, fmt.Errorf("validate attributes: %w", validationErr)
	}
	return &Event[A]{Event: ev, Attributes: attrs}, nil
}

var attributeValidator = newProjectionValidator()

func newProjectionValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		if name, _, _ := strings.Cut(field.Tag.Get(attributeFieldNameTag), ","); name != "" && name != "-" {
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
