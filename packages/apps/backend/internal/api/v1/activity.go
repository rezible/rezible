package apiv1

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type activityHandler struct{}

var activityExamples = []oapi.ActivityRecord{
	{
		Id: uuid.MustParse("bb000000-0000-4000-8000-000000000001"),
		Attributes: oapi.ActivityRecordAttributes{
			OccurredAt:  time.Date(2026, 9, 11, 8, 30, 0, 0, time.UTC),
			Explanation: "The team is monitoring recovery before closing the incident.",
			RecordKind:  "incident-update",
			RecordId:    uuid.MustParse("b1000000-0000-4000-8000-000000000003"),
			IncidentId:  new(uuid.MustParse("44444444-4444-4444-8444-444444444444")),
			Scope:       "team",
		},
	},
	{
		Id: uuid.MustParse("bb000000-0000-4000-8000-000000000002"),
		Attributes: oapi.ActivityRecordAttributes{
			OccurredAt:  time.Date(2026, 9, 10, 8, 30, 0, 0, time.UTC),
			Explanation: "Investigation completed.",
			RecordKind:  "situation-investigation",
			RecordId:    uuid.MustParse("22000000-0000-4000-8000-000000000001"),
			Scope:       "team",
		},
	},
}

func (*activityHandler) ListActivity(_ context.Context, request *oapi.ListActivityRequest) (*oapi.ListActivityResponse, error) {
	items := make([]oapi.ActivityRecord, 0)
	for _, activity := range activityExamples {
		if request.Scope != "" && activity.Attributes.Scope != request.Scope {
			continue
		}
		if !request.From.IsZero() && activity.Attributes.OccurredAt.Before(request.From) {
			continue
		}
		if !request.To.IsZero() && activity.Attributes.OccurredAt.After(request.To) {
			continue
		}
		items = append(items, activity)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Attributes.OccurredAt.Equal(items[j].Attributes.OccurredAt) {
			return items[i].Id.String() < items[j].Id.String()
		}
		return items[i].Attributes.OccurredAt.Before(items[j].Attributes.OccurredAt)
	})
	page, pageSize := request.Page, request.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	var response oapi.ListActivityResponse
	response.Body.Data = items[start:end]
	response.Body.Pagination = oapi.Pagination{Page: page, PageSize: pageSize, Total: len(items)}
	return &response, nil
}
