package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rezible/rezible/ent/incidentmilestone"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type IncidentDataProvider struct {
	apiEndpoint         string
	token               string
	webhookSecret       string
	onIncidentUpdatedFn rez.DataProviderResourceUpdatedCallback

	incidentMappingSupport *ent.Incident

	userIdEmails map[string]string
}

type IncidentDataProviderConfig struct {
	ApiEndpoint         string `json:"api_endpoint"`
	ServiceAccountToken string `json:"service_account_token"`
	WebhookSecret       string `json:"webhook_secret"`
}

func NewIncidentDataProvider(cfg IncidentDataProviderConfig) (*IncidentDataProvider, error) {
	if cfg.WebhookSecret == "" {
		return nil, fmt.Errorf("webhook secret not configured")
	}

	p := &IncidentDataProvider{
		apiEndpoint:   cfg.ApiEndpoint,
		token:         cfg.ServiceAccountToken,
		webhookSecret: cfg.WebhookSecret,
		userIdEmails:  make(map[string]string),
		onIncidentUpdatedFn: func(id string, m time.Time) {
			log.Warn().Msg("no onIncidentUpdated function")
		},
	}

	return p, nil
}

func (p *IncidentDataProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{
		"grafana_incident": http.HandlerFunc(p.makeWebhookHandler),
	}
}

func (p *IncidentDataProvider) SetOnIncidentUpdatedCallback(cb rez.DataProviderResourceUpdatedCallback) {
	p.onIncidentUpdatedFn = cb
}

func (p *IncidentDataProvider) IncidentDataMapping() *ent.Incident {
	return &incidentDataMapping
}

func (p *IncidentDataProvider) IncidentRoleDataMapping() *ent.IncidentRole {
	return &incidentRoleDataMapping
}

func (p *IncidentDataProvider) makeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	payload, parseErr := parseIncidentWebhook(r, p.webhookSecret)
	if parseErr != nil {
		http.Error(w, parseErr.Error(), http.StatusBadRequest)
		return
	}

	if payload.Incident != nil {
		incidentId := payload.Incident.IncidentID
		modifiedAt, _ := time.Parse(time.RFC3339, payload.Incident.ModifiedTime)
		if modifiedAt.IsZero() {
			modifiedAt = time.Now()
		}

		if incidentId != "" {
			go p.onIncidentUpdatedFn(incidentId, modifiedAt)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (p *IncidentDataProvider) PullIncidents(ctx context.Context) iter.Seq2[*ent.Incident, error] {
	type queryResponse struct {
		Cursor           incidentCursor    `json:"cursor"`
		Error            string            `json:"error"`
		IncidentPreviews []incidentPreview `json:"incidentPreviews"`
	}
	type queryIncidentPreviewsQuery struct {
		Limit          int    `json:"limit"`
		OrderDirection string `json:"orderDirection"`
		OrderField     string `json:"orderField"`
	}
	type queryRequest struct {
		Cursor                   *incidentCursor            `json:"cursor,omitempty"`
		IncludeCustomFieldValues bool                       `json:"includeCustomFieldValues"`
		IncludeMembershipPreview bool                       `json:"includeMembershipPreview"`
		Query                    queryIncidentPreviewsQuery `json:"query"`
	}

	return func(yield func(i *ent.Incident, err error) bool) {
		var cursor *incidentCursor
		for {
			req := queryRequest{
				Cursor: cursor,
				Query: queryIncidentPreviewsQuery{
					Limit:          25,
					OrderDirection: "DESC",
					OrderField:     "createdTime",
				},
			}
			var resp queryResponse
			respErr := p.apiRequest(ctx, "v1/IncidentsService.QueryIncidentPreviews", req, &resp)
			if respErr != nil || resp.Error != "" {
				yield(nil, fmt.Errorf("query incident previews: %w (response: %s)", respErr, resp.Error))
				return
			}
			for _, inc := range resp.IncidentPreviews {
				if !yield(convertIncidentPreview(inc), nil) {
					return
				}
			}
			if !resp.Cursor.HasMore {
				return
			}
			cursor = &resp.Cursor
		}
	}
}

func convertIncidentPreview(i incidentPreview) *ent.Incident {
	createdAt, createdErr := time.Parse(time.RFC3339, i.CreatedTime)
	modifiedAt, modifiedErr := time.Parse(time.RFC3339, i.ModifiedTime)
	closedAt, closedErr := time.Parse(time.RFC3339, i.ClosedTime)

	timeErrors := errors.Join(createdErr, modifiedErr, closedErr)
	if timeErrors != nil {
		log.Error().Err(timeErrors).Msg("failed to parse incident times")
	}

	return &ent.Incident{
		ProviderID: i.IncidentID,
		Title:      i.Title,
		Slug:       i.Slug,
		Summary:    i.Summary,
		OpenedAt:   createdAt,
		ModifiedAt: modifiedAt,
		ClosedAt:   closedAt,
	}
}

func (p *IncidentDataProvider) GetIncidentByID(ctx context.Context, id string) (*ent.Incident, error) {
	type requestBody struct {
		IncidentId string `json:"incidentID"`
	}

	req := requestBody{
		IncidentId: id,
	}
	var resp struct {
		Error    string     `json:"error"`
		Incident *gIncident `json:"incident"`
	}
	respErr := p.apiRequest(ctx, "v1/IncidentsService.GetIncident", req, &resp)
	if respErr != nil {
		return nil, respErr
	}
	if resp.Error != "" || resp.Incident == nil {
		return nil, fmt.Errorf("response error: %s", resp.Error)
	}

	return p.convertIncident(ctx, resp.Incident)
}

func (p *IncidentDataProvider) convertIncident(ctx context.Context, i *gIncident) (*ent.Incident, error) {
	createdAt, createdErr := time.Parse(time.RFC3339, i.CreatedTime)
	modifiedAt, modifiedErr := time.Parse(time.RFC3339, i.ModifiedTime)
	closedAt, closedErr := time.Parse(time.RFC3339, i.ClosedTime)

	timeErrors := errors.Join(createdErr, modifiedErr, closedErr)
	if timeErrors != nil {
		log.Error().Err(timeErrors).Msg("failed to parse incident times")
	}

	milestones, msErr := p.getIncidentMilestones(ctx, i)
	if msErr != nil {
		log.Error().Err(msErr).Msg("failed to get incident milestones")
	}

	severity := &ent.IncidentSeverity{
		Name: i.Severity,
	}

	roleAssignments, assnErr := p.convertIncidentRoleAssignments(ctx, i.IncidentMembership)
	if assnErr != nil {
		log.Error().Err(assnErr).Msg("failed to convert incident role assignments")
	}

	tags, tagsErr := p.convertIncidentLabels(ctx, i.Labels)
	if tagsErr != nil {
		log.Error().Err(tagsErr).Msg("failed to convert incident labels")
	}

	var incType *ent.IncidentType
	if i.IsDrill {
		incType = &ent.IncidentType{
			Name: "drill",
		}
	}

	tasks, tasksErr := p.convertIncidentTasks(ctx, i.TaskList)
	if tasksErr != nil {
		log.Error().Err(tasksErr).Msg("failed to convert incident tasks")
	}

	inc := &ent.Incident{
		ProviderID: i.IncidentID,
		Title:      i.Title,
		Summary:    i.Summary,
		OpenedAt:   createdAt,
		ModifiedAt: modifiedAt,
		ClosedAt:   closedAt,
		Edges: ent.IncidentEdges{
			RoleAssignments: roleAssignments,
			Severity:        severity,
			Type:            incType,
			Milestones:      milestones,
			Tasks:           tasks,
			TagAssignments:  tags,
		},
	}

	if intgErr := p.setIncidentIntegrationData(ctx, inc); intgErr != nil {
		log.Error().Err(intgErr).Msg("failed to set incident integration data")
	}

	return inc, nil
}

func (p *IncidentDataProvider) convertIncidentRoleAssignments(ctx context.Context, mem incidentMembership) ([]*ent.IncidentRoleAssignment, error) {
	var assignments []*ent.IncidentRoleAssignment
	for _, assn := range mem.Assignments {
		a := assn
		if a.User.UserID == "" { // bug in grafana irm
			continue
		}
		email, emailErr := p.lookupUserEmail(ctx, assn.User.UserID)
		if emailErr != nil {
			return nil, fmt.Errorf("failed to lookup user email: %w", emailErr)
		}
		assignments = append(assignments, &ent.IncidentRoleAssignment{
			Edges: ent.IncidentRoleAssignmentEdges{
				Role: convertIncidentRole(a.Role),
				User: &ent.User{
					Email: email,
				},
			},
		})
	}
	return assignments, nil
}

func (p *IncidentDataProvider) convertIncidentTasks(ctx context.Context, taskList incidentTaskList) ([]*ent.Task, error) {
	return nil, nil
}

func (p *IncidentDataProvider) convertIncidentLabels(ctx context.Context, labels []incidentLabel) ([]*ent.IncidentTag, error) {
	// TODO: query this using service

	tags := make([]*ent.IncidentTag, len(labels))
	for i, label := range labels {
		tags[i] = &ent.IncidentTag{
			Key:   label.Key,
			Value: label.Label,
		}
	}
	return tags, nil
}

func convertIncidentRole(r incidentRole) *ent.IncidentRole {
	var archiveTime time.Time
	if r.Archived {
		if parsed, parseErr := time.Parse(time.RFC3339, r.UpdatedAt); parseErr != nil {
			archiveTime = parsed
		}
	}
	return &ent.IncidentRole{
		Name:        r.Name,
		Required:    r.Mandatory,
		ProviderID:  strconv.Itoa(r.RoleID),
		ArchiveTime: archiveTime,
	}
}

func (p *IncidentDataProvider) getIncidentMilestones(ctx context.Context, i *gIncident) ([]*ent.IncidentMilestone, error) {
	var milestones []*ent.IncidentMilestone

	startedAt, startErr := time.Parse(time.RFC3339, i.IncidentStart)
	endedAt, endedErr := time.Parse(time.RFC3339, i.IncidentEnd)
	if timeErrs := errors.Join(startErr, endedErr); timeErrs != nil {
		return nil, fmt.Errorf("failed to parse times: %w", timeErrs)
	}

	start := &ent.IncidentMilestone{
		Type: incidentmilestone.TypeImpact,
		Time: startedAt,
	}

	end := &ent.IncidentMilestone{
		Type: incidentmilestone.TypeResolved,
		Time: endedAt,
	}

	milestones = append(milestones, start, end)

	type queryResponse struct {
		Cursor        incidentCursor         `json:"cursor"`
		Error         string                 `json:"error"`
		ActivityItems []incidentActivityItem `json:"activityItems"`
	}
	type queryActivityQuery struct {
		OrderDirection string `json:"orderDirection,omitempty"`
		Limit          int    `json:"limit,omitempty"`
		IncidentId     string `json:"incidentId"`
	}
	type queryRequest struct {
		Cursor *incidentCursor    `json:"cursor,omitempty"`
		Query  queryActivityQuery `json:"query"`
	}

	var cursor *incidentCursor
	for {
		req := queryRequest{
			Cursor: cursor,
			Query: queryActivityQuery{
				OrderDirection: "DESC",
				Limit:          100,
				IncidentId:     i.IncidentID,
			},
		}
		var resp queryResponse
		respErr := p.apiRequest(ctx, "v1/ActivityService.QueryActivity", req, &resp)
		if respErr != nil {
			return nil, fmt.Errorf("querying incident events: %w", respErr)
		}
		if resp.Error != "" {
			return nil, fmt.Errorf("querying incident events: %s", resp.Error)
		}
		for _, item := range resp.ActivityItems {
			if conv := convertIncidentMilestone(item); conv != nil {
				milestones = append(milestones, conv)
			}
		}
		if !resp.Cursor.HasMore {
			break
		}
		cursor = &resp.Cursor
	}

	return milestones, nil
}

func convertIncidentMilestone(item incidentActivityItem) *ent.IncidentMilestone {
	// fmt.Printf("incident event (%s) %s, [%s]\n", item.ActivityKind, item.CreatedTime, item.Body)

	// TODO: map properly

	event := &ent.IncidentMilestone{
		Time: item.CreatedTime,
	}

	return event
}

func (p *IncidentDataProvider) setIncidentIntegrationData(ctx context.Context, inc *ent.Incident) error {
	hookRuns, hookErr := p.getIncidentHookRuns(ctx, inc.ProviderID)
	if hookErr != nil {
		return hookErr
	}

	for _, r := range hookRuns {
		if r.HookID == "grate.irm.slack.createChannel" && inc.ChatChannelID == "" {
			parsedURL, parseErr := url.Parse(r.Metadata.Url)
			if parseErr != nil {
				return fmt.Errorf("failed to parse metadata url: %w", parseErr)
			}
			inc.ChatChannelID = parsedURL.Query().Get("channel")
			break
		}
	}

	return nil
}

func (p *IncidentDataProvider) GetRoles(ctx context.Context) ([]*ent.IncidentRole, error) {
	type responseBody struct {
		Error string         `json:"error"`
		Roles []incidentRole `json:"roles"`
	}
	var resp responseBody
	if reqErr := p.apiRequest(ctx, "RolesService.GetRoles", nil, &resp); reqErr != nil {
		return nil, fmt.Errorf("failed to list roles: %w", reqErr)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("failed to list roles resp.Error: %s", resp.Error)
	}

	roles := make([]*ent.IncidentRole, len(resp.Roles))
	for i, r := range resp.Roles {
		roles[i] = convertIncidentRole(r)
	}

	return roles, nil
}

func (p *IncidentDataProvider) getIncidentHookRuns(ctx context.Context, id string) ([]incidentHookRun, error) {
	type requestBody struct {
		IncidentId string `json:"incidentID"`
	}
	req := requestBody{IncidentId: id}

	type responseBody struct {
		Error    string            `json:"error"`
		HookRuns []incidentHookRun `json:"hookRuns"`
	}
	var resp responseBody

	if reqErr := p.apiRequest(ctx, "v1/IntegrationService.GetHookRuns", req, &resp); reqErr != nil {
		return nil, fmt.Errorf("failed to get hook runs: %w", reqErr)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("get hook runs Error: %s", resp.Error)
	}

	return resp.HookRuns, nil
}

func (p *IncidentDataProvider) lookupUserEmail(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("no id provided")
	}
	if email, exists := p.userIdEmails[id]; exists {
		return email, nil
	}

	type responseBody struct {
		Error string        `json:"error"`
		User  *incidentUser `json:"user"`
	}
	type requestBody struct {
		UserId string `json:"userID"`
	}
	var resp responseBody
	if reqErr := p.apiRequest(ctx, "v1/UsersService.GetUser", requestBody{UserId: id}, &resp); reqErr != nil {
		return "", fmt.Errorf("failed to get user: %w", reqErr)
	}
	if resp.Error != "" || resp.User == nil {
		return "", fmt.Errorf("user get error: %s", resp.Error)
	}

	email := resp.User.Email
	p.userIdEmails[id] = email
	return email, nil
}

func (p *IncidentDataProvider) apiRequest(ctx context.Context, endpoint string, reqBody any, resp any) error {
	reqUrl := fmt.Sprintf("%s/%s", p.apiEndpoint, endpoint)
	reqBytes := []byte("{}")
	if reqBody != nil {
		var bodyErr error
		reqBytes, bodyErr = json.Marshal(reqBody)
		if bodyErr != nil {
			return fmt.Errorf("failed to marshal body: %w", bodyErr)
		}
	}
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBytes))
	if reqErr != nil {
		return fmt.Errorf("failed to create request: %w", reqErr)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+p.token)

	client := &http.Client{Timeout: 60 * time.Second}

	res, resErr := client.Do(req)
	if resErr != nil {
		return fmt.Errorf("request failed: %w", resErr)
	}
	b, bErr := io.ReadAll(res.Body)
	if bErr != nil {
		return fmt.Errorf("failed to read body: %w", bErr)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", res.StatusCode)
	}
	if unmarshalErr := json.Unmarshal(b, resp); unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal: %w", unmarshalErr)
	}
	return nil
}
