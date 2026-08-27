package slackincidents

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
)

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

type sendMilestoneMessageJobArgs struct {
	IncidentId  uuid.UUID
	MilestoneId uuid.UUID
}

func (sendMilestoneMessageJobArgs) Kind() string {
	return "slackincidents-send-milestone-message"
}

func (sendMilestoneMessageJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 2,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
}

func (a *App) RegisterJobs(registry *jobs.Registry) error {
	if registerErr := registry.AddWorkerFunc(a.handleCreateIncidentChannelJob); registerErr != nil {
		return fmt.Errorf("create Slack incident channel: %w", registerErr)
	}
	if registerErr := registry.AddWorkerFunc(a.handleSendIncidentMilestoneMessageJob); registerErr != nil {
		return fmt.Errorf("send Slack incident milestone message: %w", registerErr)
	}
	return nil
}

func (a *App) handleCreateIncidentChannelJob(ctx context.Context, args createIncidentChannelJobArgs) error {
	return a.withIncidentUpdateProcessor(ctx, args.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.createIncidentChannel(ctx)
	})
}

func (a *App) handleSendIncidentMilestoneMessageJob(ctx context.Context, args sendMilestoneMessageJobArgs) error {
	return a.withIncidentUpdateProcessor(ctx, args.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.sendIncidentMilestoneMessage(ctx, args.MilestoneId)
	})
}
