package agents

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type WorkflowRunContext struct {
	Task *ent.AgentTask
	Run  *ent.AgentRun
}

func (wc WorkflowRunContext) GetSubjectEntityId(subjectKind string) (uuid.UUID, error) {
	subjects, subjectsErr := wc.Task.Edges.SubjectsOrErr()
	if subjectsErr != nil {
		return uuid.Nil, subjectsErr
	}
	for _, sub := range subjects {
		if sub.SubjectKind == subjectKind {
			if sub.DomainEntityID == nil {
				return uuid.Nil, fmt.Errorf("subject kind with nil domain entity id")
			}
			return *sub.DomainEntityID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("subject kind %s not found", subjectKind)
}

type WorkflowRunnerResult[O any] struct {
	Output   *O
	Error    error
	Metadata map[string]string
}

type Redactable interface {
	Redact() Redactable
}

func redactPayloadEntry(key string, value any) any {
	lower := strings.ToLower(key)
	if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "cookie") || strings.Contains(lower, "authorization") || strings.Contains(lower, "password") {
		return "redacted"
	}
	switch typed := value.(type) {
	case map[string]any:
		return redactMap(typed)
	case []any:
		return redactSlice(typed)
	}
	return value
}

func redactMap(inp map[string]any) map[string]any {
	redacted := make(map[string]any, len(inp))
	for k, v := range inp {
		redacted[k] = redactPayloadEntry(k, v)
	}
	return redacted
}

func redactSlice(input []any) []any {
	out := make([]any, len(input))
	for i, value := range input {
		switch typed := value.(type) {
		case map[string]any:
			out[i] = redactMap(typed)
		case []any:
			out[i] = redactSlice(typed)
		default:
			out[i] = value
		}
	}
	return out
}

var payloadValidator = newPayloadValidator()

func newPayloadValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	// validate based on json field name
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, found := strings.Cut(field.Tag.Get("json"), ",")
		if found && name != "" && name != "-" {
			return name
		}
		return field.Name
	})
	return validate
}

func validatePayload[T any](payload T) error {
	validationErr := payloadValidator.Struct(payload)
	if validationErr == nil {
		return nil
	}
	errs, isValidationErr := errors.AsType[validator.ValidationErrors](validationErr)
	if !isValidationErr || len(errs) == 0 {
		return nil
	}
	msgs := make([]string, len(errs))
	for i, verr := range errs {
		msgs[i] = fmt.Sprintf("[%s:%s]", verr.Field(), verr.Error())
	}
	return fmt.Errorf("field: %s", strings.Join(msgs, " "))
}

func encodePayload[T any](payload T) (map[string]any, error) {
	if validationErr := validatePayload(payload); validationErr != nil {
		return nil, validationErr
	}
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal payload: %w", marshalErr)
	}
	var encoded map[string]any
	if unmarshalErr := json.Unmarshal(payloadBytes, &encoded); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal encoded payload: %w", unmarshalErr)
	}
	return encoded, nil
}

func decodePayload[T any](payload map[string]any) (*T, error) {
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal payload: %w", marshalErr)
	}
	var decoded T
	decoder := json.NewDecoder(bytes.NewReader(payloadBytes))
	decoder.DisallowUnknownFields()
	if decodeErr := decoder.Decode(&decoded); decodeErr != nil {
		return nil, fmt.Errorf("decode payload: %w", decodeErr)
	}
	if validationErr := validatePayload(decoded); validationErr != nil {
		return nil, validationErr
	}
	return &decoded, nil
}
