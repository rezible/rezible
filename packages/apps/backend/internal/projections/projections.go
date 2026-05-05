package projections

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
)

type Event[T any] struct {
	Event      *ent.NormalizedEvent
	Attributes T
}

type decoder func(*ent.NormalizedEvent) (any, error)

var decoders = map[ne.Kind]decoder{
	ne.KindChatMessage:         decodeChatMessageEvent,
	ne.KindRepositoryObserved:  decodeRepositoryObservedEvent,
	ne.KindChangeEventObserved: decodeChangeEventObservedEvent,
}

func ValidateEvent(ev *ent.NormalizedEvent) (any, error) {
	if ev == nil {
		return nil, fmt.Errorf("normalized event is nil")
	}
	decode, ok := decoders[ev.Kind]
	if !ok {
		return nil, fmt.Errorf("unsupported normalized event kind %q", ev.Kind)
	}
	return decode(ev)
}

func rejectUnsupportedAttributes(ev *ent.NormalizedEvent, supported ...string) error {
	supportedSet := mapset.NewSet(supported...)
	for key := range ev.Attributes {
		if !supportedSet.Contains(key) {
			return fmt.Errorf("%s event has unsupported %s attribute", ev.Kind, key)
		}
	}
	return nil
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
