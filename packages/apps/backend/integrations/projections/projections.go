package projections

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
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

type decoder func(*ent.NormalizedEvent) (any, error)

var decoders = map[ne.Kind]decoder{
	ne.KindChatMessage:         DecodeChatMessageEvent,
	ne.KindRepositoryObserved:  DecodeRepositoryObservedEvent,
	ne.KindChangeEventObserved: DecodeChangeEventObservedEvent,
	ne.KindUserObserved:        DecodeUserObservedEvent,
	ne.KindIncidentObserved:    DecodeIncidentObservedEvent,
	ne.KindAlertObserved:       DecodeAlertObservedEvent,
}

func DecodeEvent[A any](ev *ent.NormalizedEvent) (*Event[A], error) {
	if ev == nil {
		return nil, fmt.Errorf("normalized event is nil")
	}
	decode, ok := decoders[ev.Kind]
	if !ok {
		return nil, fmt.Errorf("unsupported normalized event kind %q", ev.Kind)
	}
	decoded, decodeErr := decode(ev)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode: %w", decodeErr)
	}
	asEvent, castOk := decoded.(A)
	if !castOk {
		return nil, fmt.Errorf("failed to cast %T to T", decoded)
	}
	return &Event[A]{Event: ev, Attributes: asEvent}, nil
}

func requiredString(ev *ent.NormalizedEvent, key string) (string, error) {
	value, ok := ev.Attributes[key]
	if !ok {
		return "", fmt.Errorf("%s event missing required %s attribute", ev.Kind, key)
	}
	str, strOk := value.(string)
	if !strOk {
		return "", fmt.Errorf("%s attribute must be a string", key)
	}
	str = strings.TrimSpace(str)
	if str == "" {
		return "", fmt.Errorf("%s event missing required %s attribute", ev.Kind, key)
	}
	return str, nil
}

func optionalString(ev *ent.NormalizedEvent, key string) (string, error) {
	value, exists := ev.Attributes[key]
	if !exists {
		return "", nil
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s attribute must be a string", key)
	}
	return strings.TrimSpace(str), nil
}

func requiredInt(ev *ent.NormalizedEvent, key string) (int, error) {
	value, ok := ev.Attributes[key]
	if !ok {
		return 0, fmt.Errorf("%s event missing required %s attribute", ev.Kind, key)
	}
	switch n := value.(type) {
	case int:
		return n, nil
	case int64:
		if n < math.MinInt || n > math.MaxInt {
			return 0, fmt.Errorf("%s attribute is outside int range", key)
		}
		return int(n), nil
	case float64:
		if math.Trunc(n) != n || n < math.MinInt || n > math.MaxInt {
			return 0, fmt.Errorf("%s attribute must be an integer", key)
		}
		return int(n), nil
	default:
		return 0, fmt.Errorf("%s attribute must be an integer", key)
	}
}
