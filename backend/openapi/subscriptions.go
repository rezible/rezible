package openapi

import (
	"context"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type SubscriptionsHandler interface {
	ListSubscriptions(context.Context, *ListSubscriptionsRequest) (*ListSubscriptionsResponse, error)
	CreateSubscription(context.Context, *CreateSubscriptionRequest) (*CreateSubscriptionResponse, error)
	GetSubscription(context.Context, *GetSubscriptionRequest) (*GetSubscriptionResponse, error)
	UpdateSubscription(context.Context, *UpdateSubscriptionRequest) (*UpdateSubscriptionResponse, error)
	ArchiveSubscription(context.Context, *ArchiveSubscriptionRequest) (*ArchiveSubscriptionResponse, error)
}

func (o operations) RegisterSubscriptions(api huma.API) {
	huma.Register(api, ListSubscriptions, o.ListSubscriptions)
	huma.Register(api, CreateSubscription, o.CreateSubscription)
	huma.Register(api, GetSubscription, o.GetSubscription)
	huma.Register(api, UpdateSubscription, o.UpdateSubscription)
	huma.Register(api, ArchiveSubscription, o.ArchiveSubscription)
}

type Subscription struct {
	Id         uuid.UUID              `json:"id"`
	Attributes SubscriptionAttributes `json:"attributes"`
}

type SubscriptionAttributes struct{}

func SubscriptionFromEnt(sub *ent.Subscription) Subscription {
	return Subscription{
		Id:         sub.ID,
		Attributes: SubscriptionAttributes{},
	}
}

var subscriptionsTags = []string{"Subscriptions"}

var ListSubscriptions = huma.Operation{
	OperationID: "list-subscriptions",
	Method:      http.MethodGet,
	Path:        "/subscriptions",
	Summary:     "List Subscriptions",
	Tags:        subscriptionsTags,
	Errors:      errorCodes(),
}

type ListSubscriptionsRequest EmptyRequest
type ListSubscriptionsResponse PaginatedResponse[Subscription]

var CreateSubscription = huma.Operation{
	OperationID: "create-subscription",
	Method:      http.MethodPost,
	Path:        "/subscriptions",
	Summary:     "Create a Subscription",
	Tags:        subscriptionsTags,
	Errors:      errorCodes(),
}

type CreateSubscriptionAttributes struct {
	IncidentId *uuid.UUID `json:"incident_id" required:"false"`
	TeamId     *uuid.UUID `json:"team_id" required:"false"`
	UserId     *uuid.UUID `json:"user_id" required:"false"`
}
type CreateSubscriptionRequest RequestWithBodyAttributes[CreateSubscriptionAttributes]
type CreateSubscriptionResponse ItemResponse[Subscription]

var GetSubscription = huma.Operation{
	OperationID: "get-subscription",
	Method:      http.MethodGet,
	Path:        "/subscriptions/{id}",
	Summary:     "Get a Subscription",
	Tags:        subscriptionsTags,
	Errors:      errorCodes(),
}

type GetSubscriptionRequest GetIdRequest
type GetSubscriptionResponse ItemResponse[Subscription]

var UpdateSubscription = huma.Operation{
	OperationID: "update-subscription",
	Method:      http.MethodPatch,
	Path:        "/subscriptions/{id}",
	Summary:     "Update a Subscription",
	Tags:        subscriptionsTags,
	Errors:      errorCodes(),
}

type UpdateSubscriptionAttributes struct {
}
type UpdateSubscriptionRequest UpdateIdRequest[UpdateSubscriptionAttributes]
type UpdateSubscriptionResponse ItemResponse[Subscription]

var ArchiveSubscription = huma.Operation{
	OperationID: "archive-subscription",
	Method:      http.MethodDelete,
	Path:        "/subscriptions/{id}",
	Summary:     "Archive a Subscription",
	Tags:        subscriptionsTags,
	Errors:      errorCodes(),
}

type ArchiveSubscriptionRequest ArchiveIdRequest
type ArchiveSubscriptionResponse EmptyResponse
