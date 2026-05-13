package google

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"google.golang.org/api/meet/v2"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/videoconference"
)

var meetScopes = []string{
	"https://www.googleapis.com/auth/meetings.space.settings",
	"https://www.googleapis.com/auth/meetings.space.created",
	"https://www.googleapis.com/auth/meetings.space.readonly",
	"https://www.googleapis.com/auth/drive.meet.readonly",
}

type meetService struct {
	ci *ConfiguredIntegration
}

func newMeetService(ci *ConfiguredIntegration) *meetService {
	return &meetService{ci: ci}
}

func (s *meetService) makeClient(ctx context.Context) (*meet.Service, error) {
	creds, credsErr := s.ci.getAuthCredentials()
	if credsErr != nil {
		return nil, fmt.Errorf("failed to get auth credentials: %w", credsErr)
	}
	svc, meetErr := meet.NewService(ctx, creds)
	if meetErr != nil {
		return nil, fmt.Errorf("failed to create meet service: %w", meetErr)
	}
	return svc, nil
}

func (s *meetService) CreateVideoConference(ctx context.Context, name string) (string, error) {
	client, clientErr := s.makeClient(ctx)
	if clientErr != nil {
		return "", clientErr
	}

	createSpace := client.Spaces.Create(&meet.Space{
		Name: name,
	})

	space, reqErr := createSpace.Context(ctx).Do()
	if reqErr != nil {
		return "", fmt.Errorf("failed to create space: %w", reqErr)
	}

	return space.MeetingUri, nil
}

func (s *meetService) CreateIncidentVideoConference(ctx context.Context, inc *ent.Incident) error {
	if inc == nil {
		return fmt.Errorf("incident is nil")
	}
	exists, existsErr := inc.QueryVideoConferences().
		Where(videoconference.ProviderEQ(integrationName)).
		Where(videoconference.StatusIn(
			videoconference.StatusCreating,
			videoconference.StatusActive,
		)).
		Exist(ctx)
	if existsErr != nil {
		return fmt.Errorf("query existing incident conferences: %w", existsErr)
	}
	if exists {
		return nil
	}

	// TODO
	confName := inc.Title
	url, confErr := s.CreateVideoConference(ctx, confName)
	if confErr != nil {
		return confErr
	}

	metadata, jsonErr := json.Marshal(map[string]any{
		"provider": integrationName,
	})
	if jsonErr != nil {
		return fmt.Errorf("marshal metadata: %w", jsonErr)
	}

	setFn := func(m *ent.IncidentMutation) []ent.Mutation {
		conferenceCreate := m.Client().VideoConference.Create().
			SetIncidentID(inc.ID).
			SetProvider(integrationName).
			SetJoinURL(url).
			SetStatus(videoconference.StatusActive).
			SetMetadata(metadata).
			SetCreatedByIntegration(integrationName)
		return []ent.Mutation{conferenceCreate.Mutation()}
	}
	if _, setErr := s.ci.svcs.Incidents.Set(ctx, inc.ID, setFn); setErr != nil {
		return fmt.Errorf("set incident conference: %w", setErr)
	}
	slog.Debug("created incident conference",
		"incident_id", inc.ID.String(),
		"join_url", url,
	)
	return nil
}
