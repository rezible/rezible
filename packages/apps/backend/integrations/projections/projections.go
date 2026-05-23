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
	ne "github.com/rezible/rezible/ent/normalizedevent"

	"github.com/rezible/rezible/ent"
)

type Event[T any] struct {
	Event      *ent.NormalizedEvent
	Attributes T
}

type EventProjectionHandlerFunc = func(context.Context, *ent.Client, *ent.NormalizedEvent) error

var projectionFuncsMu sync.RWMutex
var projectionFuncs = make(map[string]EventProjectionHandlerFunc)

func RegisterHandler(name string, handler EventProjectionHandlerFunc) {
	projectionFuncsMu.Lock()
	defer projectionFuncsMu.Unlock()
	projectionFuncs[name] = handler
}

func GetHandlers() map[string]EventProjectionHandlerFunc {
	return projectionFuncs
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

func decodeKind[A any](ev *ent.NormalizedEvent, kind ne.Kind) (*Event[A], error) {
	if ev == nil {
		return nil, fmt.Errorf("normalized event is nil")
	}
	if ev.Kind != kind {
		return nil, fmt.Errorf("expected normalized event kind %q, got %q", kind, ev.Kind)
	}
	return DecodeAs[A](ev)
}

func DecodeAs[A any](ev *ent.NormalizedEvent) (*Event[A], error) {
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

func encodeAttributeStruct(attrs any) (EventAttributes, error) {
	value := reflect.ValueOf(attrs)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, fmt.Errorf("attributes are nil")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("attributes must be a struct")
	}

	typ := value.Type()
	encoded := make(EventAttributes, value.NumField())
	for i := range value.NumField() {
		fieldInfo := typ.Field(i)
		name, _, _ := strings.Cut(fieldInfo.Tag.Get(attributeFieldNameTag), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = fieldInfo.Name
		}
		encoded[name] = encodeAttributeValue(value.Field(i))
	}
	return encoded, nil
}

func encodeAttributeValue(value reflect.Value) any {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return encodeAttributeValue(value.Elem())
	}
	return value.Interface()
}
