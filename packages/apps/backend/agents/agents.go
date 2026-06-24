package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type RunResult struct {
	Content   string
	Data      map[string]any
	Citations []RunCitationInput
	Findings  []RunFindingInput
}

type TaskRunnerFunc func(context.Context, *ent.AgentTask, *ent.AgentRun) (*RunResult, error)

type RunCitationInput struct {
	CitationKind            string
	DomainEntityType        string
	DomainEntityID          uuid.UUID
	KnowledgeEntityID       uuid.UUID
	KnowledgeRelationshipID uuid.UUID
	KnowledgeEvidenceID     uuid.UUID
	AgentTaskID             uuid.UUID
	AgentRunToolCallID      uuid.UUID
	Summary                 string
	Snapshot                map[string]any
}

type RunFindingInput struct {
	FindingKind string
	Content     string
	Citations   []RunFindingCitationInput
}

type RunFindingCitationInput struct {
	CitationIndex int
	SupportKind   string
}

func FindSubjectRefID(subjects []SubjectRef, subjectType string) uuid.UUID {
	for _, subject := range subjects {
		if subject.Type == subjectType {
			return subject.ID
		}
	}
	return uuid.Nil
}

func EncodeInput[T any](input T) (map[string]any, error) {
	return encodePayload(input)
}

func DecodeInput[T any](payload map[string]any) (*T, error) {
	return decodePayload[T](payload)
}

func EncodeOutput[T any](output T) (map[string]any, error) {
	return encodePayload(output)
}

func DecodeOutput[T any](payload map[string]any) (*T, error) {
	return decodePayload[T](payload)
}

func ValidateRunCitationInput(input RunCitationInput) error {
	if strings.TrimSpace(input.CitationKind) == "" {
		return fmt.Errorf("agent citation kind is required")
	}
	if strings.TrimSpace(input.Summary) == "" {
		return fmt.Errorf("agent citation summary is required")
	}
	targets := 0
	if input.DomainEntityType != "" || input.DomainEntityID != uuid.Nil {
		if input.DomainEntityType == "" || input.DomainEntityID == uuid.Nil {
			return fmt.Errorf("domain citation requires type and id")
		}
		targets++
	}
	for _, id := range []uuid.UUID{input.KnowledgeEntityID, input.KnowledgeRelationshipID, input.KnowledgeEvidenceID, input.AgentTaskID, input.AgentRunToolCallID} {
		if id != uuid.Nil {
			targets++
		}
	}
	if targets != 1 {
		return fmt.Errorf("agent citation requires exactly one target")
	}
	return nil
}

func RedactPayload(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "cookie") || strings.Contains(lower, "authorization") || strings.Contains(lower, "password") {
			out[key] = "redacted"
			continue
		}
		switch typed := value.(type) {
		case map[string]any:
			out[key] = RedactPayload(typed)
		case []any:
			out[key] = redactSlice(typed)
		default:
			out[key] = value
		}
	}
	return out
}

func redactSlice(input []any) []any {
	out := make([]any, len(input))
	for i, value := range input {
		switch typed := value.(type) {
		case map[string]any:
			out[i] = RedactPayload(typed)
		case []any:
			out[i] = redactSlice(typed)
		default:
			out[i] = value
		}
	}
	return out
}

var payloadValidator = newPayloadValidator()

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

func newPayloadValidator() *validator.Validate {
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

func validatePayload[T any](payload T) error {
	if validationErr := payloadValidator.Struct(payload); validationErr != nil {
		if errs, ok := errors.AsType[validator.ValidationErrors](validationErr); ok && len(errs) > 0 {
			msgs := make([]string, len(errs))
			for i, verr := range errs {
				msgs[i] = fmt.Sprintf("[%s:%s]", verr.Field(), verr.Error())
			}
			return fmt.Errorf("field: %s", strings.Join(msgs, " "))
		}
		return validationErr
	}
	return nil
}
