package slackincidents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
)

// syncIncidentChannelJobArgs brings an incident's response channel up to date at DueAt, now rounded up to the
// second. The job is unique by args, so changes within a second are one sync, and a change while a sync runs
// schedules another rather than being dropped.
type syncIncidentChannelJobArgs struct {
	IncidentID uuid.UUID `json:"incident_id" river:"unique"`
	DueAt      time.Time `json:"due_at" river:"unique"`
}

func (syncIncidentChannelJobArgs) Kind() string {
	return "slackincidents-sync-incident-channel"
}

func (syncIncidentChannelJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 5,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
}

type createIncidentChannelJobArgs struct {
	IncidentId uuid.UUID
}

func (createIncidentChannelJobArgs) Kind() string {
	return "slackincidents-create-incident-channel"
}

func (createIncidentChannelJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 2,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
}

type postIncidentMilestoneJobArgs struct {
	IncidentID  uuid.UUID
	MilestoneID uuid.UUID
}

func (postIncidentMilestoneJobArgs) Kind() string {
	return "slackincidents-post-incident-milestone"
}

func (postIncidentMilestoneJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 2,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
}

func (a *App) JobWorkers() []jobs.WorkerDefinition {
	return []jobs.WorkerDefinition{
		jobs.DefineWorkerFunc(a.handleSyncIncidentChannelJob),
		jobs.DefineWorkerFunc(a.handleCreateIncidentChannelJob),
		jobs.DefineWorkerFunc(a.handlePostIncidentMilestoneJob),
	}
}

func (a *App) handleSyncIncidentChannelJob(ctx context.Context, args syncIncidentChannelJobArgs) error {
	return a.syncIncidentChannel(ctx, args.IncidentID)
}

func (a *App) handleCreateIncidentChannelJob(ctx context.Context, args createIncidentChannelJobArgs) error {
	return a.createIncidentChannel(ctx, args.IncidentId)
}

func (a *App) handlePostIncidentMilestoneJob(ctx context.Context, args postIncidentMilestoneJobArgs) error {
	return a.postIncidentMilestone(ctx, args.IncidentID, args.MilestoneID)
}
