package jobs

import "github.com/google/uuid"

type SendIncidentDebriefRequestsJobArgs struct {
	IncidentId uuid.UUID
}

func (SendIncidentDebriefRequestsJobArgs) Kind() string {
	return "send-incident-debrief-requests"
}
