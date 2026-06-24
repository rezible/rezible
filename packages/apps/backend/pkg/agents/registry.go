package agents

import (
	"sync"
)

type WorkflowRegistry struct {
	runners   map[WorkflowKind]TaskRunnerFunc
	runnersMu sync.RWMutex
}

func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{runners: make(map[WorkflowKind]TaskRunnerFunc)}
}

func (r *WorkflowRegistry) Register(kind WorkflowKind, runner TaskRunnerFunc) {
	r.runnersMu.Lock()
	defer r.runnersMu.Unlock()
	r.runners[kind] = runner
}

func (r *WorkflowRegistry) Get(kind WorkflowKind) (TaskRunnerFunc, bool) {
	if r == nil {
		return nil, false
	}
	r.runnersMu.RLock()
	defer r.runnersMu.RUnlock()
	runner, ok := r.runners[kind]
	return runner, ok
}
