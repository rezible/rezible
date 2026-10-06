package demoprovider

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/projections"
)

func TestDemoIncidentAlertsAreRecentAndOneKeepsFiring(t *testing.T) {
	installedAt := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	latencyResolved := false
	for _, event := range demoAlertEventsAt(installedAt) {
		if event.OccurredAt.After(installedAt) || event.receivedAt().After(installedAt) {
			t.Fatalf("%s lands in the future at %s", event.NotificationRef, event.OccurredAt)
		}
		isIncidentAlert := event.DefinitionRef == demoSearchLatencyFiring.DefinitionRef ||
			event.DefinitionRef == demoElasticsearchCPUFiring.DefinitionRef
		if isIncidentAlert && installedAt.Sub(event.StartedAt) > time.Hour {
			t.Fatalf("incident alert %s started %s before install", event.NotificationRef, installedAt.Sub(event.StartedAt))
		}
		if event.DefinitionRef == demoSearchLatencyFiring.DefinitionRef && event.State == projections.AlertStateResolved {
			latencyResolved = true
		}
	}
	if latencyResolved {
		t.Fatalf("the search latency alert resolves, so its situation would close")
	}
}

func TestDemoAlertsAreTheSameOnEveryPull(t *testing.T) {
	installedAt := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	installed := &InstalledIntegration{intg: &ent.Integration{CreatedAt: installedAt}}
	querier := newEventQuerier(installed)
	pullAlerts := func(cursors rez.ProviderEventSourceCursors) map[string]rez.ProviderEvent {
		alerts := make(map[string]rez.ProviderEvent)
		for result, pullErr := range querier.QueryProviderEvents(t.Context(), cursors) {
			if pullErr != nil {
				t.Fatalf("pull demo events: %v", pullErr)
			}
			if result.Event.ProviderEventSource == sourceAlerts {
				alerts[result.Event.ProviderEventRef] = result.Event
			}
		}
		return alerts
	}
	startedAt := func(alert rez.ProviderEvent) time.Time {
		var payload alertObservedPayload
		if decodeErr := json.Unmarshal(alert.Attributes, &payload); decodeErr != nil {
			t.Fatalf("decode demo alert %s: %v", alert.ProviderEventRef, decodeErr)
		}
		return payload.StartedAt
	}

	all := pullAlerts(rez.ProviderEventSourceCursors{})
	// A later pull resumes after the CPU alert fired, before its resolve was delivered.
	resumeAfter := rez.ProviderEventSourceCursors{sourceAlerts: demoElasticsearchCPUFiring.cursorAfter()}
	resumed := pullAlerts(resumeAfter)

	if len(resumed) == 0 || len(resumed) >= len(all) {
		t.Fatalf("the resumed pull delivered %d of %d alerts; want some but not all", len(resumed), len(all))
	}
	for ref, later := range resumed {
		first, delivered := all[ref]
		if !delivered {
			t.Fatalf("the resumed pull delivered %s, which the first pull did not", ref)
		}
		if !bytes.Equal(first.Attributes, later.Attributes) || !first.ReceivedAt.Equal(later.ReceivedAt) {
			t.Fatalf("%s differs between pulls", ref)
		}
	}

	firing := all[demoElasticsearchCPUFiring.eventRef()]
	resolved, delivered := resumed[resolvedDemoAlert(demoElasticsearchCPUFiring, demoIncidentAlertsEnd).eventRef()]
	if !delivered {
		t.Fatalf("the resumed pull did not deliver the CPU alert's resolve")
	}
	if !startedAt(firing).Equal(startedAt(resolved)) {
		t.Fatalf("the CPU alert's resolve started at %s, its firing at %s", startedAt(resolved), startedAt(firing))
	}
	if !resolved.ReceivedAt.Equal(installedAt.Add(-time.Minute)) {
		t.Fatalf("the last incident alert notification was received at %s; want a minute before install", resolved.ReceivedAt)
	}
}
