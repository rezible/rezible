package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
)

type OncallDataProvider struct {
	apiEndpoint string
	apiToken    string

	providerUserMap map[string]*ent.User
}

type OncallDataProviderConfig struct {
	ApiEndpoint string `json:"api_endpoint"`
	ApiToken    string `json:"api_token"`
}

func NewOncallDataProvider(cfg OncallDataProviderConfig) (*OncallDataProvider, error) {
	p := &OncallDataProvider{
		apiEndpoint:     strings.TrimSuffix(cfg.ApiEndpoint, "/"),
		apiToken:        cfg.ApiToken,
		providerUserMap: make(map[string]*ent.User),
	}

	go func() {
		ctx := context.Background()
		if userErr := p.updateUserData(ctx); userErr != nil {
			log.Error().Err(userErr).Msg("failed to update oncall user data")
		}
	}()

	return p, nil
}

func (p *OncallDataProvider) RosterDataMapping() *ent.OncallRoster {
	return &rosterMapping
}

func (p *OncallDataProvider) UserShiftDataMapping() *ent.OncallUserShift {
	return &shiftMapping
}

func (p *OncallDataProvider) PullRosters(ctx context.Context) iter.Seq2[*ent.OncallRoster, error] {
	return func(yield func(*ent.OncallRoster, error) bool) {
		if usersErr := p.updateUserData(ctx); usersErr != nil {
			yield(nil, fmt.Errorf("updating user data: %w", usersErr))
			return
		}

		schedules, listSchedulesErr := p.listSchedules(ctx)
		if listSchedulesErr != nil {
			yield(nil, fmt.Errorf("querying oncall schedules (rosters): %w", listSchedulesErr))
			return
		}

		shifts, listShiftsErr := p.listShifts(ctx)
		if listShiftsErr != nil {
			yield(nil, fmt.Errorf("querying oncall shifts (schedules): %w", listShiftsErr))
			return
		}

		shiftIdMap := make(map[string]*oncallShift)
		for _, shift := range shifts {
			s := shift
			shiftIdMap[s.Id] = &s
		}

		for _, schedule := range schedules {
			if !yield(p.convertSchedule(schedule, shiftIdMap), nil) {
				return
			}
		}
	}
}

func (p *OncallDataProvider) ListRosters(ctx context.Context) ([]*ent.OncallRoster, error) {
	if usersErr := p.updateUserData(ctx); usersErr != nil {
		return nil, fmt.Errorf("updating user data: %w", usersErr)
	}

	schedules, listSchedulesErr := p.listSchedules(ctx)
	if listSchedulesErr != nil {
		return nil, fmt.Errorf("list oncall schedules: %w", listSchedulesErr)
	}

	shifts, listShiftsErr := p.listShifts(ctx)
	if listShiftsErr != nil {
		return nil, fmt.Errorf("list oncall shifts: %w", listShiftsErr)
	}

	shiftIdMap := make(map[string]*oncallShift)
	for _, shift := range shifts {
		s := shift
		shiftIdMap[s.Id] = &s
	}

	rosters := make([]*ent.OncallRoster, len(schedules))
	for i, schedule := range schedules {
		rosters[i] = p.convertSchedule(schedule, shiftIdMap)
	}

	return rosters, nil
}

func (p *OncallDataProvider) FetchOncallersForRoster(ctx context.Context, rosterId string) ([]*ent.User, error) {
	reqUrl := oncallApiUrl(p.apiEndpoint, "schedules/"+rosterId, nil)
	var resp oncallSchedule
	if getErr := oncallGet(ctx, reqUrl, p.apiToken, &resp); getErr != nil {
		return nil, fmt.Errorf("get schedules request: %w", getErr)
	}
	users := make([]*ent.User, len(resp.OnCallNow))
	for i, id := range resp.OnCallNow {
		usr, ok := p.providerUserMap[id]
		if !ok {
			return nil, fmt.Errorf("no user found")
		}
		users[i] = usr
	}
	return users, nil
}

func (p *OncallDataProvider) PullShiftsForRoster(ctx context.Context, id string, from, to time.Time) iter.Seq2[*ent.OncallUserShift, error] {
	formatFrom := formatOncallTime(from.UTC())
	formatTo := formatOncallTime(to.UTC())

	params := &url.Values{}
	params.Set("start_date", formatFrom)
	params.Set("end_date", formatTo)
	endpoint := fmt.Sprintf("schedules/%s/final_shifts", id)
	initialUrl := oncallApiUrl(p.apiEndpoint, endpoint, params)

	isIncompleteShift := func(shiftStart, shiftEnd time.Time) bool {
		if formatFrom == formatOncallTime(shiftStart.UTC()) {
			return true
		}
		return formatTo == formatOncallTime(shiftEnd.UTC())
	}

	return func(yield func(*ent.OncallUserShift, error) bool) {
		reqUrl := &initialUrl
		for reqUrl != nil {
			var resp oncallPaginatedResponse[oncallUserShift]
			if getErr := oncallGet(ctx, *reqUrl, p.apiToken, &resp); getErr != nil {
				yield(nil, getErr)
				return
			}
			for _, res := range resp.Results {
				user, userExists := p.providerUserMap[res.UserPk]
				if !userExists {
					user = &ent.User{
						Name:  res.UserUsername,
						Email: res.UserEmail,
					}
				}
				startsAt, startErr := time.Parse(time.RFC3339, res.ShiftStart)
				if startErr != nil {
					log.Error().Err(startErr).Str("time", res.ShiftStart).Msg("failed to parse shift start time")
				}
				endsAt, endErr := time.Parse(time.RFC3339, res.ShiftEnd)
				if endErr != nil {
					log.Error().Err(endErr).Str("time", res.ShiftEnd).Msg("failed to parse shift end time")
				}

				if isIncompleteShift(startsAt, endsAt) {
					log.Debug().Msg("skipping incomplete shift")
					continue
				}

				shift := &ent.OncallUserShift{
					ProviderID: fmt.Sprintf("%s_%s_%s_%s", id, res.UserPk, res.ShiftStart, res.ShiftEnd),
					StartAt:    startsAt,
					EndAt:      endsAt,
					Edges: ent.OncallUserShiftEdges{
						User: user,
					},
				}

				if !yield(shift, nil) {
					return
				}
			}
			reqUrl = resp.Next
		}
	}
}

func (p *OncallDataProvider) ListShiftsForRoster(ctx context.Context, id string, from, to time.Time) ([]*ent.OncallUserShift, error) {
	params := &url.Values{}
	params.Set("start_date", formatOncallTime(from))
	params.Set("end_date", formatOncallTime(to))
	endpoint := fmt.Sprintf("schedules/%s/final_shifts", id)
	reqUrl := oncallApiUrl(p.apiEndpoint, endpoint, params)

	var shifts []*ent.OncallUserShift
	for {
		var resp oncallPaginatedResponse[oncallUserShift]
		if getErr := oncallGet(ctx, reqUrl, p.apiToken, &resp); getErr != nil {
			return nil, fmt.Errorf("get schedules request: %w", getErr)
		}
		for _, res := range resp.Results {
			user, userExists := p.providerUserMap[res.UserPk]
			if !userExists {
				user = &ent.User{
					Name:  res.UserUsername,
					Email: res.UserEmail,
				}
			}
			startsAt, startErr := time.Parse(time.RFC3339, res.ShiftStart)
			if startErr != nil {
				log.Error().Err(startErr).Str("time", res.ShiftStart).Msg("failed to parse shift start time")
			}
			endsAt, endErr := time.Parse(time.RFC3339, res.ShiftEnd)
			if endErr != nil {
				log.Error().Err(endErr).Str("time", res.ShiftEnd).Msg("failed to parse shift end time")
			}
			shifts = append(shifts, &ent.OncallUserShift{
				StartAt: startsAt,
				EndAt:   endsAt,
				Edges: ent.OncallUserShiftEdges{
					User: user,
				},
			})
		}
		if resp.Next == nil {
			break
		}
		reqUrl = *resp.Next
	}

	return shifts, nil
}

func (p *OncallDataProvider) updateUserData(ctx context.Context) error {
	newUserMap := make(map[string]*ent.User)
	reqUrl := oncallApiUrl(p.apiEndpoint, "users", nil)
	for {
		var resp oncallPaginatedResponse[oncallUser]
		if getErr := oncallGet(ctx, reqUrl, p.apiToken, &resp); getErr != nil {
			return fmt.Errorf("list users request: %w", getErr)
		}
		for _, user := range resp.Results {
			u := user
			newUserMap[u.Id] = &ent.User{
				Name:  u.Username,
				Email: u.Email,
				// Timezone: u.Timezone
			}
		}
		if resp.Next == nil {
			break
		}
		reqUrl = *resp.Next
	}
	p.providerUserMap = newUserMap
	return nil
}

func (p *OncallDataProvider) convertSchedule(s oncallSchedule, shiftIdMap map[string]*oncallShift) *ent.OncallRoster {
	var rosterSchedules []*ent.OncallSchedule
	for _, shiftId := range s.Shifts {
		shift, exists := shiftIdMap[shiftId]
		if !exists {
			log.Warn().Str("id", shiftId).Msg("no matching shift found during conversion")
			continue
		}
		rosterSchedules = append(rosterSchedules, p.convertShift(shift))
	}

	return &ent.OncallRoster{
		Name:          s.Name,
		ProviderID:    s.Id,
		Timezone:      s.TimeZone,
		ChatChannelID: s.Slack.ChannelId,
		ChatHandle:    s.Slack.UserGroupId,
		Edges: ent.OncallRosterEdges{
			Schedules: rosterSchedules,
		},
	}
}

func (p *OncallDataProvider) convertShift(shift *oncallShift) *ent.OncallSchedule {
	// TODO
	sched := &ent.OncallSchedule{
		ProviderID: shift.Id,
		Name:       shift.Name,
	}

	if shift.TimeZone != nil {
		sched.Timezone = *shift.TimeZone
	}

	addParticipant := func(index int, providerId string) {
		user, ok := p.providerUserMap[providerId]
		if !ok {
			log.Warn().Str("providerId", providerId).Msg("unknown provider user")
			return
		}
		sched.Edges.Participants = append(sched.Edges.Participants, &ent.OncallScheduleParticipant{
			Index: index,
			Edges: ent.OncallScheduleParticipantEdges{
				User: user,
			},
		})
	}

	for idx, userId := range shift.Users {
		addParticipant(idx, userId)
	}

	for idx, users := range shift.RollingUsers {
		for _, userId := range users {
			addParticipant(idx, userId)
		}
	}

	return sched
}

func (p *OncallDataProvider) listShifts(ctx context.Context) ([]oncallShift, error) {
	var results []oncallShift
	reqUrl := oncallApiUrl(p.apiEndpoint, "on_call_shifts", nil)
	for {
		var resp oncallPaginatedResponse[oncallShift]
		if getErr := oncallGet(ctx, reqUrl, p.apiToken, &resp); getErr != nil {
			return nil, fmt.Errorf("get schedules request: %w", getErr)
		}
		if resp.Count > 0 {
			results = append(results, resp.Results...)
		}
		if resp.Next == nil {
			break
		}
		reqUrl = *resp.Next
	}

	return results, nil
}

func (p *OncallDataProvider) listSchedules(ctx context.Context) ([]oncallSchedule, error) {
	var results []oncallSchedule
	reqUrl := oncallApiUrl(p.apiEndpoint, "schedules", nil)
	for {
		var resp oncallPaginatedResponse[oncallSchedule]
		if getErr := oncallGet(ctx, reqUrl, p.apiToken, &resp); getErr != nil {
			return nil, fmt.Errorf("get schedules request: %w", getErr)
		}
		if resp.Count > 0 {
			results = append(results, resp.Results...)
		}
		if resp.Next == nil {
			break
		}
		reqUrl = *resp.Next
	}

	return results, nil
}

func oncallGet(ctx context.Context, reqUrl string, token string, resp any) error {
	body, resErr := oncallRequest(ctx, http.MethodGet, reqUrl, token, nil)
	if resErr != nil {
		return fmt.Errorf("request failed: %w", resErr)
	}

	if unmarshalErr := json.Unmarshal(body, resp); unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal: %w", unmarshalErr)
	}

	return nil
}

func oncallRequest(ctx context.Context, method string, url string, apiToken string, body io.Reader) ([]byte, error) {
	req, reqErr := http.NewRequestWithContext(ctx, method, url, body)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", apiToken)

	res, resErr := http.DefaultClient.Do(req)
	if resErr != nil {
		return nil, fmt.Errorf("request failed: %w", resErr)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", res.StatusCode)
	}

	b, bErr := io.ReadAll(res.Body)
	if bErr != nil {
		return nil, fmt.Errorf("failed to read body: %w", bErr)
	}

	return b, nil
}

func formatOncallTime(t time.Time) string {
	return t.Format("2006-01-02T15:04")
}

func oncallApiUrl(base, endpoint string, values *url.Values) string {
	apiUrl := fmt.Sprintf("%s/api/v1/%s", base, endpoint)
	if values != nil {
		apiUrl = apiUrl + "?" + values.Encode()
	}
	return apiUrl
}
