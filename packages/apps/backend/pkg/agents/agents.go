package agents

import (
	"fmt"

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
