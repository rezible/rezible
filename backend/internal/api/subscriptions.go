package api

import (
	"context"

	oapi "github.com/twohundreds/rezible/openapi"
)

type subscriptionsHandler struct {
}

func newSubscriptionsHandler() *subscriptionsHandler {
	return &subscriptionsHandler{}
}

func (h *subscriptionsHandler) ListSubscriptions(ctx context.Context, input *oapi.ListSubscriptionsRequest) (*oapi.ListSubscriptionsResponse, error) {
	var resp oapi.ListSubscriptionsResponse

	return &resp, nil
}

func (h *subscriptionsHandler) CreateSubscription(ctx context.Context, input *oapi.CreateSubscriptionRequest) (*oapi.CreateSubscriptionResponse, error) {
	var resp oapi.CreateSubscriptionResponse

	return &resp, nil
}

func (h *subscriptionsHandler) GetSubscription(ctx context.Context, input *oapi.GetSubscriptionRequest) (*oapi.GetSubscriptionResponse, error) {
	var resp oapi.GetSubscriptionResponse

	return &resp, nil
}

func (h *subscriptionsHandler) UpdateSubscription(ctx context.Context, input *oapi.UpdateSubscriptionRequest) (*oapi.UpdateSubscriptionResponse, error) {
	var resp oapi.UpdateSubscriptionResponse

	return &resp, nil
}

func (h *subscriptionsHandler) ArchiveSubscription(ctx context.Context, input *oapi.ArchiveSubscriptionRequest) (*oapi.ArchiveSubscriptionResponse, error) {
	var resp oapi.ArchiveSubscriptionResponse

	return &resp, nil
}
