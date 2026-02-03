package google

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/meet/v2"
	"google.golang.org/api/option"
)

var meetScopes = []string{
	"https://www.googleapis.com/auth/meetings.space.settings",
	"https://www.googleapis.com/auth/meetings.space.created",
	"https://www.googleapis.com/auth/meetings.space.readonly",
	"https://www.googleapis.com/auth/drive.meet.readonly",
}

const PreferenceEnableIncidentVideoConferences = "incident_video_conferences"

type meetService struct {
	msgs     rez.MessageService
	credsOpt option.ClientOption
}

func newMeetService(ctx context.Context, msgs rez.MessageService, credsOpt option.ClientOption) (*meetService, error) {
	return &meetService{
		msgs:     msgs,
		credsOpt: credsOpt,
	}, nil
}

func (s *meetService) makeClient(ctx context.Context) (*meet.Service, error) {
	svc, meetErr := meet.NewService(ctx, s.credsOpt)
	if meetErr != nil {
		return nil, fmt.Errorf("failed to create meet service: %w", meetErr)
	}
	return svc, nil
}

func (s *meetService) CreateVideoConference(ctx context.Context, inc *ent.Incident) (string, error) {
	client, clientErr := s.makeClient(ctx)
	if clientErr != nil {
		return "", clientErr
	}

	createSpace := client.Spaces.Create(&meet.Space{
		Name: "",
	})

	space, reqErr := createSpace.Context(ctx).Do()
	if reqErr != nil {
		return "", fmt.Errorf("failed to create space: %w", reqErr)
	}

	return space.MeetingUri, nil
}

func (s *meetService) CreateIncidentVideoConference(ctx context.Context, inc *ent.Incident) error {
	log.Debug().Msg("meet service.. create incident conference")
	return nil
}
