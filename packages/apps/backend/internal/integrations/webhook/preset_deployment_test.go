package webhook

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
)

var (
	testInstallationID = uuid.MustParse("0b6f3c1e-8d2a-4f5b-9c7e-2a1d3e4f5a6b")
	testReceivedAt     = time.Date(2026, 10, 7, 14, 4, 0, 0, time.UTC)
)

type DeploymentPresetSuite struct {
	test.Suite
}

func TestDeploymentPresetSuite(t *testing.T) {
	suite.Run(t, &DeploymentPresetSuite{Suite: test.NewSuite()})
}

// testReport is the spec's example report.
func testReport() map[string]any {
	return map[string]any{
		"id":          "8812345678-1",
		"service":     "checkout-api",
		"environment": "production",
		"status":      "succeeded",
		"repository":  "acme/checkout",
		"sha":         "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392",
		"version":     "v1.42.0",
		"url":         "https://github.com/acme/checkout/actions/runs/8812345678",
		"started_at":  "2026-10-07T14:01:10Z",
		"finished_at": "2026-10-07T14:03:52Z",
	}
}

// withFields returns the example report with fields changed, removing those set to nil.
func withFields(changes map[string]any) map[string]any {
	report := testReport()
	for field, value := range changes {
		if value == nil {
			delete(report, field)
		} else {
			report[field] = value
		}
	}
	return report
}

func encodeReport(t *testing.T, report map[string]any) []byte {
	t.Helper()
	encoded, encodeErr := json.Marshal(report)
	if encodeErr != nil {
		t.Fatalf("encode report: %v", encodeErr)
	}
	return encoded
}

func (s *DeploymentPresetSuite) mapDelivery(body []byte) (*rez.ProviderEvent, error) {
	return deploymentPreset{}.MapDelivery(testInstallationID, body, testReceivedAt)
}

// mapAndProcess maps a report and normalizes its provider event through the event processor.
func (s *DeploymentPresetSuite) mapAndProcess(report map[string]any) (*rez.ProviderEvent, *projections.DeploymentEvent) {
	s.T().Helper()
	event, mapErr := s.mapDelivery(encodeReport(s.T(), report))
	s.Require().NoError(mapErr)

	normalized, processErr := EventProcessor{}.ProcessProviderEvent(s.T().Context(), *event)
	s.Require().NoError(processErr)
	s.Require().Len(normalized, 1)
	s.Require().Equal(projections.KindDeployment, normalized[0].Kind)

	decoded, decodeErr := projections.DecodeDeploymentEvent(normalized[0])
	s.Require().NoError(decodeErr)
	return event, decoded
}

func (s *DeploymentPresetSuite) requireRejected(report map[string]any, field string) {
	s.T().Helper()
	_, mapErr := s.mapDelivery(encodeReport(s.T(), report))
	var rejection *fieldError
	s.Require().True(errors.As(mapErr, &rejection), "rejected with a field error: %v", mapErr)
	s.Require().Equal(field, rejection.field, rejection.Error())
	s.Require().True(strings.HasPrefix(rejection.Error(), field+": "))
}

func (s *DeploymentPresetSuite) TestMapsAReport() {
	report := testReport()
	report["service"] = "  Checkout_API "
	report["environment"] = "Production"
	report["repository"] = "Acme/Checkout"
	report["sha"] = "4F1C0D9E8B7A6F5E4D3C2B1A09F8E7D6C5B4A392"
	report["pipeline"] = "ignored"

	event, decoded := s.mapAndProcess(report)
	attrs := decoded.Attributes

	s.Equal(ProviderName, event.Provider)
	s.Equal(testInstallationID.String(), event.ProviderNamespace)
	s.Equal(presetDeployment, event.ProviderEventSource, "the event source is the preset's name")
	s.Equal(testReceivedAt, event.ReceivedAt)

	s.Equal("8812345678-1", attrs.ExternalID)
	s.Equal("succeeded", attrs.Status)
	expectedService := projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          ProviderName,
			ProviderNamespace: testInstallationID.String(),
			ResourceRef:       "service:checkout-api",
		},
		Category:    kne.CategoryContainer,
		Kind:        "service",
		DisplayName: "Checkout_API",
		LinkingAttributes: projections.LinkingAttributes{
			"service.name": "checkout-api",
		},
	}
	s.Equal(expectedService, attrs.Service)
	expectedEnvironment := projections.DeploymentEnvironment{
		Name:        "production",
		DisplayName: "Production",
	}
	s.Equal(expectedEnvironment, attrs.Environment)
	expectedRepository := &projections.EntityObservation{
		Ref: rez.ProviderResourceRef{
			Provider:          ProviderName,
			ProviderNamespace: testInstallationID.String(),
			ResourceRef:       "repository:acme/checkout",
		},
		Category:    kne.CategoryCode,
		Kind:        "repository",
		DisplayName: "Acme/Checkout",
		LinkingAttributes: projections.LinkingAttributes{
			"repository.full_name": "acme/checkout",
		},
	}
	s.Equal(expectedRepository, attrs.Repository)
	s.Equal("4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392", attrs.Sha)
	s.Equal("v1.42.0", attrs.Version)
	s.Equal("https://github.com/acme/checkout/actions/runs/8812345678", attrs.URL)
	s.Equal(time.Date(2026, 10, 7, 14, 1, 10, 0, time.UTC), *attrs.StartedAt)
	s.Equal(time.Date(2026, 10, 7, 14, 3, 52, 0, time.UTC), *attrs.FinishedAt)
}

func (s *DeploymentPresetSuite) TestAcceptsFieldRules() {
	maxName := strings.Repeat("a", maxNameLength)
	cases := map[string]map[string]any{
		"started":                      {"status": "started", "finished_at": nil},
		"failed":                       {"status": "failed"},
		"no optional fields":           {"repository": nil, "sha": nil, "version": nil, "url": nil, "started_at": nil, "finished_at": nil},
		"id of 200 characters":         {"id": maxName},
		"service of 200 characters":    {"service": maxName},
		"environment of 200":           {"environment": maxName},
		"version of 200 characters":    {"version": maxName},
		"repository punctuation":       {"repository": "acme.io/check_out-api.v2"},
		"repository of 100 per part":   {"repository": strings.Repeat("o", 100) + "/" + strings.Repeat("n", 100)},
		"http url":                     {"url": "http://ci.internal/runs/1"},
		"url of 2000 characters":       {"url": "https://ci.test/" + strings.Repeat("r", maxURLLength-len("https://ci.test/"))},
		"finished when started":        {"started_at": "2026-10-07T14:01:10Z", "finished_at": "2026-10-07T14:01:10Z"},
		"started before defaulted end": {"started_at": "2026-10-07T14:04:00Z", "finished_at": nil},
		"times with offsets":           {"started_at": "2026-10-08T00:01:10+10:00", "finished_at": "2026-10-07T14:03:52.5Z"},
		"time 5 minutes after now":     {"finished_at": "2026-10-07T14:09:00Z"},
		"null optional fields":         {"repository": json.RawMessage("null"), "sha": json.RawMessage("null")},
		"whitespace around required":   {"id": "  8812345678-1  ", "status": " succeeded "},
		"unknown fields are no error":  {"deployer": map[string]any{"name": "ci"}},
	}
	for name, changes := range cases {
		s.Run(name, func() {
			_, mapErr := s.mapDelivery(encodeReport(s.T(), withFields(changes)))
			s.Require().NoError(mapErr)
		})
	}
}

func (s *DeploymentPresetSuite) TestRejectsFieldRules() {
	tooLong := strings.Repeat("a", maxNameLength+1)
	cases := map[string]struct {
		changes map[string]any
		field   string
	}{
		"missing id":                  {map[string]any{"id": nil}, "id"},
		"empty id":                    {map[string]any{"id": "  "}, "id"},
		"null id":                     {map[string]any{"id": json.RawMessage("null")}, "id"},
		"numeric id":                  {map[string]any{"id": 8812345678}, "id"},
		"id too long":                 {map[string]any{"id": tooLong}, "id"},
		"missing service":             {map[string]any{"service": nil}, "service"},
		"empty service":               {map[string]any{"service": ""}, "service"},
		"service without a name":      {map[string]any{"service": "---"}, "service"},
		"service too long":            {map[string]any{"service": tooLong}, "service"},
		"missing environment":         {map[string]any{"environment": nil}, "environment"},
		"environment without a name":  {map[string]any{"environment": "__"}, "environment"},
		"environment too long":        {map[string]any{"environment": tooLong}, "environment"},
		"missing status":              {map[string]any{"status": nil}, "status"},
		"unknown status":              {map[string]any{"status": "success"}, "status"},
		"repository without owner":    {map[string]any{"repository": "checkout"}, "repository"},
		"repository with two slashes": {map[string]any{"repository": "acme/checkout/api"}, "repository"},
		"repository with empty part":  {map[string]any{"repository": "acme/"}, "repository"},
		"repository with a space":     {map[string]any{"repository": "acme/check out"}, "repository"},
		"repository part too long":    {map[string]any{"repository": "acme/" + strings.Repeat("n", 101)}, "repository"},
		"repository URL":              {map[string]any{"repository": "https://github.com/acme/checkout"}, "repository"},
		"short sha":                   {map[string]any{"sha": "4f1c0d9"}, "sha"},
		"sha that is not hexadecimal": {map[string]any{"sha": strings.Repeat("g", 40)}, "sha"},
		"version too long":            {map[string]any{"version": tooLong}, "version"},
		"url without http":            {map[string]any{"url": "ftp://ci.test/runs/1"}, "url"},
		"url without a host":          {map[string]any{"url": "https:///runs/1"}, "url"},
		"relative url":                {map[string]any{"url": "/runs/1"}, "url"},
		"url too long":                {map[string]any{"url": "https://ci.test/" + strings.Repeat("r", maxURLLength)}, "url"},
		"started_at not RFC3339":      {map[string]any{"started_at": "2026-10-07 14:01:10"}, "started_at"},
		"finished_at not RFC3339":     {map[string]any{"finished_at": "yesterday"}, "finished_at"},
		"finished before started":     {map[string]any{"finished_at": "2026-10-07T14:01:09Z"}, "finished_at"},
		"started_at in the future":    {map[string]any{"started_at": "2026-10-07T14:09:01Z", "finished_at": nil}, "started_at"},
		"finished_at in the future":   {map[string]any{"finished_at": "2026-10-07T14:09:01Z"}, "finished_at"},
		"finished_at on a started":    {map[string]any{"status": "started"}, "finished_at"},
		"numeric finished_at":         {map[string]any{"finished_at": 1}, "finished_at"},
		// Within the skew allowance, but after the receive time that stands in for the missing finished_at.
		"succeeded starting after receipt": {map[string]any{"status": "succeeded", "started_at": "2026-10-07T14:05:00Z", "finished_at": nil}, "started_at"},
		"failed starting after receipt":    {map[string]any{"status": "failed", "started_at": "2026-10-07T14:05:00Z", "finished_at": nil}, "started_at"},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			s.requireRejected(withFields(tc.changes), tc.field)
		})
	}
}

func (s *DeploymentPresetSuite) TestRejectsBodiesThatAreNotObjects() {
	for _, body := range []string{"", "not json", "null", `["id"]`, `"report"`} {
		_, mapErr := s.mapDelivery([]byte(body))

		var rejection *fieldError
		s.Require().True(errors.As(mapErr, &rejection), "body %q", body)
		s.Equal("body", rejection.field)
	}
}

func (s *DeploymentPresetSuite) TestTreatsEmptyOptionalFieldsAsAbsent() {
	report := withFields(map[string]any{
		"repository": "",
		"sha":        " ",
		"version":    "",
		"url":        "",
		"started_at": "",
	})

	_, decoded := s.mapAndProcess(report)

	attrs := decoded.Attributes
	s.Nil(attrs.Repository)
	s.Empty(attrs.Sha)
	s.Empty(attrs.Version)
	s.Empty(attrs.URL)
	s.Nil(attrs.StartedAt)
}

func (s *DeploymentPresetSuite) TestDefaultsMissingTimes() {
	s.Run("started", func() {
		report := withFields(map[string]any{"status": "started", "started_at": nil, "finished_at": nil})

		_, decoded := s.mapAndProcess(report)

		s.Equal(testReceivedAt, *decoded.Attributes.StartedAt)
		s.Nil(decoded.Attributes.FinishedAt)
	})

	for _, status := range []string{"succeeded", "failed"} {
		s.Run(status, func() {
			report := withFields(map[string]any{"status": status, "started_at": nil, "finished_at": nil})

			_, decoded := s.mapAndProcess(report)

			s.Nil(decoded.Attributes.StartedAt)
			s.Equal(testReceivedAt, *decoded.Attributes.FinishedAt)
		})
	}
}

func (s *DeploymentPresetSuite) TestEventTimeIsTheDeploymentTime() {
	_, finished := s.mapAndProcess(testReport())
	_, started := s.mapAndProcess(withFields(map[string]any{"status": "started", "finished_at": nil}))

	s.Equal(time.Date(2026, 10, 7, 14, 3, 52, 0, time.UTC), finished.Event.OccurredAt, "a finished deployment's time is finished_at")
	s.Equal(time.Date(2026, 10, 7, 14, 1, 10, 0, time.UTC), started.Event.OccurredAt, "an unfinished deployment's time is started_at")
}

func (s *DeploymentPresetSuite) TestDeploymentAndReportIdentity() {
	refs := func(changes map[string]any) (string, string) {
		_, decoded := s.mapAndProcess(withFields(changes))
		return decoded.Event.ProviderResourceRef, decoded.Event.ProviderEventRef
	}

	deployment, succeededReport := refs(nil)
	s.Equal(deploymentRef("checkout-api", "production", "8812345678-1"), deployment)
	s.True(strings.HasPrefix(deployment, "deployment:"))
	s.Equal(deployment+":succeeded", succeededReport)

	aliased, aliasedReport := refs(map[string]any{"service": "Checkout_API", "environment": " PRODUCTION "})
	s.Equal(deployment, aliased, "Checkout_API and checkout-api with the same ID are one deployment")
	s.Equal(succeededReport, aliasedReport, "a redelivered report has the same identity")

	changed, changedReport := refs(map[string]any{"version": "v1.43.0", "sha": nil})
	s.Equal(deployment, changed)
	s.Equal(succeededReport, changedReport, "a report with the same status is a duplicate even if other fields differ")

	started, startedReport := refs(map[string]any{"status": "started", "finished_at": nil})
	s.Equal(deployment, started, "reports of one deployment share its ref")
	s.Equal(deployment+":started", startedReport)

	otherEnvironment, _ := refs(map[string]any{"environment": "staging"})
	otherID, _ := refs(map[string]any{"id": "8812345678-2"})
	upperCaseID, _ := refs(map[string]any{"id": "RUN-1"})
	lowerCaseID, _ := refs(map[string]any{"id": "run-1"})
	s.NotEqual(deployment, otherEnvironment)
	s.NotEqual(deployment, otherID)
	s.NotEqual(upperCaseID, lowerCaseID, "the caller's ID is used as sent")
}
