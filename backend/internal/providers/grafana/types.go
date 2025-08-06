package grafana

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const mb = 1 << (10 * 2)

func parseIncidentWebhook(r *http.Request, signingSecret string) (*incidentWebhookPayload, error) {
	if verifyErr := verifyIncidentWebhookSignature(r, signingSecret); verifyErr != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", verifyErr)
	}

	payload, readErr := io.ReadAll(io.LimitReader(r.Body, int64(1*mb)))
	if readErr != nil {
		return nil, fmt.Errorf("failed to read body: %w", readErr)
	}

	var payloadObject incidentWebhookPayload
	if jsonErr := json.Unmarshal(payload, &payloadObject); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal webhook payload: %w", jsonErr)
	}
	return &payloadObject, nil
}

func verifyIncidentWebhookSignature(r *http.Request, signingSecret string) error {
	header := r.Header["Gi-Signature"]
	if len(header) == 0 || header[0] == "" {
		return errors.New("empty GI-Signature")
	}

	signatures := strings.Split(header[0], ",")
	s := make(map[string]string)
	for _, pair := range signatures {
		values := strings.Split(pair, "=")
		s[values[0]] = values[1]
	}
	t := s["t"]
	v1 := s["v1"]

	payload, readErr := io.ReadAll(io.LimitReader(r.Body, int64(1*mb)))

	// Copy body for other handlers
	r.Body = io.NopCloser(bytes.NewBuffer(payload))

	if readErr != nil {
		return fmt.Errorf("failed to read body: %w", readErr)
	}

	hasher := sha256.New()
	hasher.Write(payload)
	bodyHash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	stringToSign := bodyHash + ":" + t + ":v1"

	m := hmac.New(sha256.New, []byte(signingSecret))
	m.Write([]byte(stringToSign))
	expectedHash := hex.EncodeToString(m.Sum(nil))

	if expectedHash != v1 {
		return errors.New("invalid GI-Signature")
	}

	return nil
}

// grafana-incident go sdk types are unreliable/out of date
type (
	incidentCursor struct {
		// NextValue is the start position of the next set of results. The implementation
		// may change, so clients should not rely on this value.
		NextValue string `json:"nextValue"`

		// HasMore indicates whether there are more results or not. If HasMore is true,
		// you can make the same request again (except using this Cursor instead) to get
		// the next page of results.
		HasMore bool `json:"hasMore"`
	}
	incidentUserPreview struct {
		UserID   string `json:"userID"`
		Name     string `json:"name"`
		PhotoURL string `json:"photoURL"`
	}

	gIncident struct {
		IncidentID         string              `json:"incidentID"`
		Title              string              `json:"title"`
		Summary            string              `json:"summary"`
		Severity           string              `json:"severity"`
		Status             string              `json:"status"`
		IsDrill            bool                `json:"isDrill"`
		Labels             []incidentLabel     `json:"labels"`
		IncidentStart      string              `json:"incidentStart"`
		CreatedTime        string              `json:"createdTime"`
		DurationSeconds    int                 `json:"durationSeconds"`
		IncidentEnd        string              `json:"incidentEnd"`
		ClosedTime         string              `json:"closedTime"`
		ModifiedTime       string              `json:"modifiedTime"`
		CreatedByUser      incidentUserPreview `json:"createdByUser"`
		IncidentMembership incidentMembership  `json:"incidentMembership"`
		TaskList           incidentTaskList    `json:"taskList"`
		HeroImagePath      string              `json:"heroImagePath"`
		OverviewURL        string              `json:"overviewURL"`
	}

	incidentPreview struct {
		ClosedTime    string              `json:"closedTime"`
		CreatedByUser incidentUserPreview `json:"createdByUser"`
		CreatedTime   string              `json:"createdTime"`
		Description   string              `json:"description"`
		FieldValues   []struct {
			FieldUUID string `json:"fieldUUID"`
			Value     string `json:"value"`
		} `json:"fieldValues"`
		HeroImagePath             string `json:"heroImagePath"`
		IncidentEnd               string `json:"incidentEnd"`
		IncidentID                string `json:"incidentID"`
		IncidentMembershipPreview struct {
			ImportantAssignments []struct {
				RoleID int                 `json:"roleID"`
				User   incidentUserPreview `json:"user"`
			} `json:"importantAssignments"`
			TotalAssignments  int `json:"totalAssignments"`
			TotalParticipants int `json:"totalParticipants"`
		} `json:"incidentMembershipPreview"`
		IncidentStart string          `json:"incidentStart"`
		IncidentType  string          `json:"incidentType"`
		IsDrill       bool            `json:"isDrill"`
		Labels        []incidentLabel `json:"labels"`
		ModifiedTime  string          `json:"modifiedTime"`
		SeverityID    string          `json:"severityID"`
		SeverityLabel string          `json:"severityLabel"`
		Slug          string          `json:"slug"`
		Status        string          `json:"status"`
		Summary       string          `json:"summary"`
		Title         string          `json:"title"`
		Version       int             `json:"version"`
	}

	incidentLabel struct {
		ColorHex    string `json:"colorHex"`
		Description string `json:"description"`
		Key         string `json:"key"`
		Label       string `json:"label"`
	}

	incidentMembership struct {
		Assignments       []incidentRoleAssignment `json:"assignments"`
		TotalAssignments  int                      `json:"totalAssignments"`
		TotalParticipants int                      `json:"totalParticipants"`
	}

	incidentRole struct {
		Archived    bool   `json:"archived"`
		CreatedAt   string `json:"createdAt"`
		Description string `json:"description"`
		Important   bool   `json:"important"`
		Mandatory   bool   `json:"mandatory"`
		Name        string `json:"name"`
		OrgID       string `json:"orgID"`
		RoleID      int    `json:"roleID"`
		UpdatedAt   string `json:"updatedAt"`
	}

	incidentRoleAssignment struct {
		Role   incidentRole        `json:"role"`
		RoleID int                 `json:"roleID"`
		User   incidentUserPreview `json:"user"`
	}

	incidentTaskList struct {
		Tasks     []incidentTask `json:"tasks"`
		DoneCount int            `json:"doneCount"`
		TodoCount int            `json:"todoCount"`
	}

	incidentTask struct {
		AssignedUser incidentUserPreview `json:"assignedUser"`
		AuthorUser   incidentUserPreview `json:"authorUser"`
		CreatedTime  string              `json:"createdTime"`
		Immutable    bool                `json:"immutable"`
		ModifiedTime string              `json:"modifiedTime"`
		Status       string              `json:"status"`
		TaskID       string              `json:"taskID"`
		Text         string              `json:"text"`
	}

	incidentUser struct {
		Email          string `json:"email"`
		GrafanaLogin   string `json:"grafanaLogin"`
		GrafanaUserID  string `json:"grafanaUserID"`
		InternalUserID string `json:"internalUserID"`
		ModifiedTime   string `json:"modifiedTime"`
		MsTeamsUserID  string `json:"msTeamsUserID"`
		Name           string `json:"name"`
		PhotoURL       string `json:"photoURL"`
		SlackTeamID    string `json:"slackTeamID"`
		SlackUserID    string `json:"slackUserID"`
		UserID         string `json:"userID"`
	}

	incidentHookRun struct {
		IntegrationID string                  `json:"integrationID"`
		HookID        string                  `json:"hookID"`
		EnabledHookID string                  `json:"enabledHookID"`
		LastRun       string                  `json:"lastRun"`
		LastUpdate    string                  `json:"lastUpdate"`
		Metadata      incidentHookRunMetadata `json:"metadata"`
		EventName     string                  `json:"eventName"`
		EventKind     string                  `json:"eventKind"`
		UpdateStatus  string                  `json:"updateStatus"`
		UpdateError   string                  `json:"updateError"`
		Status        string                  `json:"status"`
		Error         string                  `json:"error"`
	}

	incidentHookRunMetadata struct {
		Title       string `json:"title"`
		Explanation string `json:"explanation"`
		Url         string `json:"url"`
	}

	incidentActivityItem struct {
		ActivityItemID string                           `json:"activityItemID"`
		ActivityKind   string                           `json:"activityKind"`
		Attachments    []incidentActivityItemAttachment `json:"attachments"`
		Body           string                           `json:"body"`
		CreatedTime    time.Time                        `json:"createdTime"`
		EventTime      time.Time                        `json:"eventTime"`
		FieldValues    map[string]any                   `json:"fieldValues"`
		Immutable      bool                             `json:"immutable"`
		IncidentID     string                           `json:"incidentID"`
		Relevance      string                           `json:"relevance"`
		SubjectUser    incidentUser                     `json:"subjectUser"`
		Tags           []string                         `json:"tags"`
		Url            string                           `json:"url"`
		User           incidentUser                     `json:"user"`
	}

	incidentActivityItemAttachment struct {
		AttachedByUserID string    `json:"attachedByUserID"`
		AttachmentErr    string    `json:"attachmentErr"`
		AttachmentID     string    `json:"attachmentID"`
		ContentLength    int       `json:"contentLength"`
		ContentType      string    `json:"contentType"`
		DeletedTime      time.Time `json:"deletedTime"`
		DisplayType      string    `json:"displayType"`
		DownloadURL      string    `json:"downloadURL"`
		Ext              string    `json:"ext"`
		FileType         string    `json:"fileType"`
		HasThumbnail     bool      `json:"hasThumbnail"`
		Path             string    `json:"path"`
		SHA512           string    `json:"sHA512"`
		SourceURL        string    `json:"sourceURL"`
		ThumbnailURL     string    `json:"thumbnailURL"`
		UploadTime       time.Time `json:"uploadTime"`
		UseSourceURL     bool      `json:"useSourceURL"`
	}

	incidentWebhookPayload struct {
		ID       string     `json:"id"`
		Version  string     `json:"version"`
		Source   string     `json:"source"`
		Time     string     `json:"time"`
		Event    string     `json:"event"`
		Incident *gIncident `json:"incident"`
	}
)

type (
	oncallPaginatedResponse[T any] struct {
		Count             int     `json:"count"`
		Next              *string `json:"next"`
		Previous          *string `json:"previous"`
		Results           []T     `json:"results"`
		CurrentPageNumber int     `json:"current_page_number"`
		PageSize          int     `json:"page_size"`
		TotalPages        int     `json:"total_pages"`
	}

	oncallTeam struct {
		Id        string `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarUrl string `json:"avatar_url"`
	}

	oncallUser struct {
		Id    string `json:"id"`
		Email string `json:"email"`
		Slack struct {
			UserId string `json:"user_id"`
			TeamId string `json:"team_id"`
		} `json:"slack"`
		Username string   `json:"username"`
		Role     string   `json:"role"`
		Timezone string   `json:"timezone"`
		TeamIds  []string `json:"teams"`
	}

	oncallShift struct {
		Id                     string     `json:"id"`
		Name                   string     `json:"name"`
		TeamId                 *string    `json:"team_id"`
		TimeZone               *string    `json:"time_zone"` // defaults to schedule
		Level                  int        `json:"level"`
		Start                  string     `json:"start"`
		Duration               int        `json:"duration"`
		Type                   string     `json:"type"`      // single_event,recurrent_event,rolling_users
		Frequency              string     `json:"frequency"` // if !single_event: hourly,daily,weekly,monthly
		Interval               int        `json:"interval"`
		Until                  string     `json:"until"` // format yyyy-MM-dd'T'HH:mm:ss
		WeekStart              string     `json:"week_start"`
		ByDay                  []string   `json:"by_day"`
		ByMonth                []int      `json:"by_month"`
		ByMonthDay             []int      `json:"by_monthday"`
		Users                  []string   `json:"users"`
		RollingUsers           [][]string `json:"rolling_users"`
		StartRotationFromIndex int        `json:"start_rotation_from_user_index"`
	}

	OncallShift struct {
		UserPk       string `json:"user_pk"`
		UserEmail    string `json:"user_email"`
		UserUsername string `json:"user_username"`
		ShiftStart   string `json:"shift_start"`
		ShiftEnd     string `json:"shift_end"`
	}

	oncallSchedule struct {
		Id               string   `json:"id"`
		Name             string   `json:"name"`
		Type             string   `json:"type"`
		TeamId           *string  `json:"team_id"`
		TimeZone         string   `json:"time_zone"`
		OnCallNow        []string `json:"on_call_now"`
		Shifts           []string `json:"shifts"`
		IcalUrlPrimary   *string  `json:"ical_url_primary"`
		IcalUrlOverrides *string  `json:"ical_url_overrides"`
		Slack            struct {
			ChannelId   string `json:"channel_id"`
			UserGroupId string `json:"user_group_id"`
		} `json:"slack"`
	}

	alertGroup struct {
		Id             string    `json:"id"`
		IntegrationId  string    `json:"integration_id"`
		RouteId        string    `json:"route_id"`
		AlertsCount    int       `json:"alerts_count"`
		State          string    `json:"state"`
		CreatedAt      time.Time `json:"created_at"`
		ResolvedAt     time.Time `json:"resolved_at"`
		AcknowledgedAt *string   `json:"acknowledged_at,omitempty"`
		AcknowledgedBy *string   `json:"acknowledged_by,omitempty"`
		ResolvedBy     string    `json:"resolved_by"`
		Title          string    `json:"title"`
		Permalinks     struct {
			Slack    string `json:"slack"`
			Telegram string `json:"telegram"`
		} `json:"permalinks"`
		SilencedAt time.Time      `json:"silenced_at"`
		LastAlert  *alertInstance `json:"last_alert"`
	}

	alertInstance struct {
		Id           string                `json:"id"`
		AlertGroupId string                `json:"alert_group_id"`
		CreatedAt    time.Time             `json:"created_at"`
		Payload      *alertInstancePayload `json:"payload"`
	}

	alertInstancePayload struct {
		State       string `json:"state"`
		Title       string `json:"title"`
		RuleId      int    `json:"ruleId"`
		Message     string `json:"message"`
		RuleUrl     string `json:"ruleUrl"`
		RuleName    string `json:"ruleName"`
		EvalMatches []struct {
			Tags   interface{} `json:"tags"`
			Value  int         `json:"value"`
			Metric string      `json:"metric"`
		} `json:"evalMatches"`
	}
)
