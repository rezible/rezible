package projections

import (
	"bytes"
	"encoding/json"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

// AttributePath is a path into a change point's event attributes, one segment per JSON key, separated by "/".
// A segment may contain dots, as linking attribute keys do.
type AttributePath = rez.AttributePath

// The involvement roles: the assertion of the evidence tying a change point's event to an entity it involves.
const (
	AssertionDeploymentService    = "deployment_service"
	AssertionDeploymentRepository = "deployment_repository"
	AssertionChangeRepository     = "change_repository"
	AssertionChangeImpacted       = "change_impacted"
)

// InvolvementAssertions returns the involvement roles of an event kind, the role that wins when an entity
// fills two first. A kind that records no change point has none.
func InvolvementAssertions(kind string) []string {
	switch kind {
	case KindDeployment:
		return []string{AssertionDeploymentService, AssertionDeploymentRepository}
	case KindCodeChange:
		return []string{AssertionChangeRepository, AssertionChangeImpacted}
	default:
		return nil
	}
}

// ChangePointEventAfter reports whether the incoming event sorts after the current event of the same change
// point, each at its change point time: by later time, then for deployments a terminal status over started,
// then by greater event ID. Event IDs are UUIDv7, so a greater ID was recorded later. An event never sorts
// after itself.
func ChangePointEventAfter(incoming *ent.NormalizedEvent, incomingAt time.Time, current *ent.NormalizedEvent, currentAt time.Time) bool {
	if !incomingAt.Equal(currentAt) {
		return incomingAt.After(currentAt)
	}
	if incoming.Kind == KindDeployment {
		incomingFinished, currentFinished := deploymentEventFinished(incoming), deploymentEventFinished(current)
		if incomingFinished != currentFinished {
			return incomingFinished
		}
	}
	return bytes.Compare(incoming.ID[:], current.ID[:]) > 0
}

// deploymentEventFinished reports whether a deployment event's status is terminal. Stored events were
// validated when normalized, so an undecodable one counts as not finished.
func deploymentEventFinished(event *ent.NormalizedEvent) bool {
	var attributes DeploymentEventAttributes
	if json.Unmarshal(event.Attributes, &attributes) != nil {
		return false
	}
	return attributes.Finished()
}
