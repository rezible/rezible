package projections

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/rezible/rezible/ent"
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

type EventAttributes map[string]any

func EncodeAttributes[A any](attrs A) (EventAttributes, error) {
	if validationErr := validateAttributes(attrs); validationErr != nil {
		return nil, fmt.Errorf("validate attributes: %w", validationErr)
	}
	attrBytes, marshalErr := json.Marshal(attrs)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal attributes: %w", marshalErr)
	}
	var encodedAttrs EventAttributes
	if unmarshalErr := json.Unmarshal(attrBytes, &encodedAttrs); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal encoded attributes: %w", unmarshalErr)
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
		MatchName:   matchAttributeName,
		DecodeHook:  projectionAttributeDecodeHook(),
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

func matchAttributeName(mapKey, fieldName string) bool {
	mapKey = strings.ReplaceAll(mapKey, "_", "")
	fieldName = strings.ReplaceAll(fieldName, "_", "")
	return strings.EqualFold(mapKey, fieldName)
}

func projectionAttributeDecodeHook() mapstructure.DecodeHookFunc {
	timeType := reflect.TypeOf(time.Time{})
	return mapstructure.ComposeDecodeHookFunc(
		func(from reflect.Type, to reflect.Type, data any) (any, error) {
			if to != timeType {
				return data, nil
			}
			if from.Kind() == reflect.Map {
				value := reflect.ValueOf(data)
				if value.Len() == 0 {
					return time.Time{}, nil
				}
			}
			return data, nil
		},
		mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
	)
}

func newProjectionValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, found := strings.Cut(field.Tag.Get(attributeFieldNameTag), ",")
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
