package google

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"google.golang.org/api/meet/v2"
	"google.golang.org/api/option"
)

type meetService struct {
	msgs         rez.MessageService
	integrations rez.IntegrationsService
}

func newMeetService(ctx context.Context, svcs *rez.Services) (*meetService, error) {
	return &meetService{
		msgs:         svcs.Messages,
		integrations: svcs.Integrations,
	}, nil
}

var meetScopes = []string{
	"https://www.googleapis.com/auth/meetings.space.settings",
	"https://www.googleapis.com/auth/meetings.space.created",
	"https://www.googleapis.com/auth/meetings.space.readonly",
	"https://www.googleapis.com/auth/drive.meet.readonly",
}

func (s *meetService) getClient(ctx context.Context) (*meet.Service, error) {
	cfg, cfgErr := lookupIntegrationConfig(ctx, s.integrations)
	if cfgErr != nil {
		return nil, cfgErr
	}

	credsOpt := option.WithAuthCredentialsJSON(option.ServiceAccount, cfg.UserConfig.ServiceAccountCredentials)
	client, meetErr := meet.NewService(ctx, credsOpt)
	if meetErr != nil {
		return nil, fmt.Errorf("failed to create meet service: %w", meetErr)
	}

	return client, nil
}

func (s *meetService) CreateVideoConference(ctx context.Context) (string, error) {
	client, clientErr := s.getClient(ctx)
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
