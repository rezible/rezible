package google

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/videoconference"
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
	incidents rez.IncidentService
	credsOpt  option.ClientOption
}

func newMeetService(ctx context.Context, incidents rez.IncidentService, credsOpt option.ClientOption) (*meetService, error) {
	return &meetService{
		incidents: incidents,
		credsOpt:  credsOpt,
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
	curr, currErr := s.incidents.Get(ctx, inc.ID)
	if currErr != nil {
		return fmt.Errorf("get incident: %w", currErr)
	}

	exists, existsErr := curr.QueryVideoConferences().
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

	url, confErr := s.CreateVideoConference(ctx, curr)
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
		incidentId, exists := m.ID()
		if !exists {
			return nil
		}
		conferenceCreate := m.Client().VideoConference.Create().
			SetIncidentID(incidentId).
			SetProvider(integrationName).
			SetJoinURL(url).
			SetStatus(videoconference.StatusActive).
			SetMetadata(metadata).
			SetCreatedByIntegration(integrationName)
		return []ent.Mutation{conferenceCreate.Mutation()}
	}
	if _, setErr := s.incidents.Set(ctx, curr.ID, setFn); setErr != nil {
		return fmt.Errorf("set incident conference: %w", setErr)
	}
	log.Debug().Str("incident_id", curr.ID.String()).Str("join_url", url).Msg("created incident conference")
	return nil
}
