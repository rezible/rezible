package projections

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

type Event[T any] struct {
	Event      *ent.NormalizedEvent
	Attributes T
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

func DecodeEventAttributes[A any](ev *ent.NormalizedEvent) (*Event[A], error) {
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

type EntityObservation struct {
	Ref         rez.ProviderResourceRef `json:"ref" validate:"required"`
	Category    kne.Category            `json:"category" validate:"required"`
	Kind        string                  `json:"kind" validate:"required"`
	DisplayName string                  `json:"display_name" validate:"required"`
	Description string                  `json:"description,omitempty"`
	Properties  map[string]any          `json:"properties,omitempty"`
}

func SortEntityObservations(observations []EntityObservation) []EntityObservation {
	sorted := append([]EntityObservation(nil), observations...)
	slices.SortStableFunc(sorted, func(left, right EntityObservation) int {
		if left.Ref.Provider != right.Ref.Provider {
			return strings.Compare(left.Ref.Provider, right.Ref.Provider)
		}
		if left.Ref.ProviderNamespace != right.Ref.ProviderNamespace {
			return strings.Compare(left.Ref.ProviderNamespace, right.Ref.ProviderNamespace)
		}
		if left.Ref.ResourceRef != right.Ref.ResourceRef {
			return strings.Compare(left.Ref.ResourceRef, right.Ref.ResourceRef)
		}
		if left.Category != right.Category {
			return strings.Compare(left.Category.String(), right.Category.String())
		}
		if left.Kind != right.Kind {
			return strings.Compare(left.Kind, right.Kind)
		}
		return strings.Compare(left.DisplayName, right.DisplayName)
	})
	return sorted
}

func DerivedRelationshipRef(predicate knr.Predicate, source, target rez.ProviderResourceRef) rez.ProviderResourceRef {
	digest := sha256.New()
	writeString := func(value string) {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(value)))
		digest.Write(length[:])
		digest.Write([]byte(value))
	}

	writeString("relationship:v1")
	writeString(predicate.String())
	writeString(source.Provider)
	writeString(source.ProviderNamespace)
	writeString(source.ResourceRef)
	writeString(target.Provider)
	writeString(target.ProviderNamespace)
	writeString(target.ResourceRef)

	return rez.ProviderResourceRef{
		Provider:    "rezible",
		ResourceRef: fmt.Sprintf("relationship:v1:%x", digest.Sum(nil)),
	}
}
