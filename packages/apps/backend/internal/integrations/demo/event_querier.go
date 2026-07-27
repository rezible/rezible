package demoprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/integrations"
)

func (i *Integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	return newEventQuerier(&InstalledIntegration{intg: intg}), nil
}

type eventQuerier struct {
	ii *InstalledIntegration
}

func newEventQuerier(ci *InstalledIntegration) *eventQuerier {
	return &eventQuerier{ii: ci}
}

func (q *eventQuerier) QueryProviderEvents(ctx context.Context, cursors rez.ProviderEventQuerySourceCursors) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	demoComponents := makeDemoTopologyComponents()
	demoRelationships := makeDemoTopologyRelationships(demoComponents)
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		pullFuncs := []func() bool{
			makeEventPuller(cursors, yield, sourceUsers, demoUserEvents),
			makeEventPuller(cursors, yield, sourceCodeRepos, demoCodeRepositoryEvents),
			makeEventPuller(cursors, yield, sourceCodeChanges, demoCodeChangeEvents),
			makeEventPuller(cursors, yield, sourceAlerts, demoAlertEvents),
			makeEventPuller(cursors, yield, sourceIncidents, demoIncidentEvents),
			makeEventPuller(cursors, yield, sourceTopology, demoComponents),
			makeEventPuller(cursors, yield, sourceTopology, demoRelationships),
		}
		for _, pullFunc := range pullFuncs {
			if !pullFunc() {
				return
			}
		}
	}
}

func makeEventPuller[P demoEventPayload](cursors rez.ProviderEventQuerySourceCursors, yield func(*rez.ProviderEventQueryResult, error) bool, source string, payloads []P) func() bool {
	return func() bool {
		if cursor, shouldQuery := integrations.GetSourceQueryCursor(cursors, source); shouldQuery {
			for ev, evErr := range pullPayloadEvents(source, payloads, cursor) {
				if !yield(ev, evErr) {
					return false
				}
			}
		}
		return true
	}
}

func pullPayloadEvents[P demoEventPayload](source string, items []P, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		for _, p := range items {
			cursorAfter := p.subjectRef()
			if cursor != "" && cursorAfter <= cursor {
				continue
			}
			enc, jsonErr := json.Marshal(p)
			if jsonErr != nil {
				yield(nil, fmt.Errorf("marshal demo event: %w", jsonErr))
				return
			}
			ev := rez.ProviderEvent{
				Provider:           integrationName,
				ProviderSource:     source,
				ProviderEventRef:   p.subjectRef(),
				ProviderSubjectRef: p.subjectRef(),
				ReceivedAt:         demoObservedAt,
				Payload:            enc,
				ContentType:        "application/json",
			}
			res := &rez.ProviderEventQueryResult{Event: ev, SourceCursorAfter: new(cursorAfter)}
			if !yield(res, nil) {
				return
			}
		}
	}
}
