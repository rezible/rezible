package apiv1

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	kpd "github.com/rezible/rezible/ent/knowledgeproposaldecision"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/errs"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test/mocks"
)

// knowledgeFixture is a service named `checkout` by one alias and `checkout-api` by another, which a pending
// proposal says may be one.
type knowledgeFixture struct {
	recordedAt  time.Time
	member      *ent.User
	checkout    *ent.KnowledgeEntity
	checkoutAPI *ent.KnowledgeEntity
	proposal    *ent.KnowledgeProposal
	rejected    *ent.KnowledgeProposal
}

func newKnowledgeFixture() *knowledgeFixture {
	f := &knowledgeFixture{
		recordedAt: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		member:     &ent.User{ID: uuid.New(), Name: "Member"},
	}
	f.checkout = &ent.KnowledgeEntity{
		ID:       uuid.New(),
		Category: kne.CategoryContainer,
		Kind:     "service",
		State:    schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout"},
	}
	f.checkout.Edges.LinkingAttributes = []*ent.KnowledgeEntityLinkingAttribute{
		{Attribute: "service.name", Value: "checkout"},
	}
	f.checkoutAPI = &ent.KnowledgeEntity{
		ID:       uuid.New(),
		Category: kne.CategoryContainer,
		Kind:     "service",
		State:    schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout-api"},
	}
	evidenceAlias := &ent.KnowledgeSubjectAlias{
		ID:                  uuid.New(),
		Provider:            "provider-b",
		ProviderResourceRef: "deployment:checkout",
		EntityID:            &f.checkout.ID,
	}
	evidence := &ent.KnowledgeEvidence{
		ID:           uuid.New(),
		Assertion:    "deployment_observed",
		EffectiveAt:  f.recordedAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout"},
	}
	evidence.Edges.SubjectAlias = evidenceAlias
	f.proposal = &ent.KnowledgeProposal{
		ID:        uuid.New(),
		Category:  kp.CategoryContainer,
		Kind:      "service",
		Attribute: "service.name",
		ValueA:    "checkout",
		ValueB:    "checkout-api",
		Status:    kp.StatusPending,
		Origin:    kp.OriginNameSimilarity,
		Rationale: "`checkout` and `checkout-api` differ only by the generic term `api`, so they may name the same service.",
	}
	f.proposal.Edges.Evidence = evidence
	f.rejected = &ent.KnowledgeProposal{
		ID:        uuid.New(),
		Attribute: "service.name",
		ValueA:    "checkout",
		ValueB:    "checkout-svc",
		Status:    kp.StatusRejected,
	}
	rejection := &ent.KnowledgeProposalDecision{Status: kpd.StatusRejected, CreatedAt: f.recordedAt}
	rejection.Edges.User = f.member
	f.rejected.Edges.Decisions = []*ent.KnowledgeProposalDecision{rejection}
	return f
}

// withAliases is an entity with one alias, as the knowledge services load it.
func (f *knowledgeFixture) withAliases(entity *ent.KnowledgeEntity, provider, integration, name string) *rez.KnowledgeEntityWithAliases {
	alias := rez.KnowledgeEntityAlias{
		Alias:                  &ent.KnowledgeSubjectAlias{ID: uuid.New(), SubjectKind: ksa.SubjectKindEntity, Provider: provider, ProviderResourceRef: "service:" + name},
		IntegrationDisplayName: integration,
		Name:                   name,
		LastRecordedAt:         f.recordedAt,
	}
	return &rez.KnowledgeEntityWithAliases{Entity: entity, Aliases: []rez.KnowledgeEntityAlias{alias}}
}

func (f *knowledgeFixture) proposalWithValues() *rez.KnowledgeProposalWithValues {
	return &rez.KnowledgeProposalWithValues{
		Proposal: f.proposal,
		ValueA: rez.KnowledgeProposalValue{
			Value:  "checkout",
			Holder: f.withAliases(f.checkout, "provider-b", "Provider B", "checkout"),
		},
		ValueB: rez.KnowledgeProposalValue{
			Value:  "checkout-api",
			Holder: f.withAliases(f.checkoutAPI, "provider-a", "Provider A", "checkout-api"),
		},
		EvidenceIntegrationDisplayName: "Provider B",
	}
}

func TestListKnowledgeProposals(t *testing.T) {
	f := newKnowledgeFixture()
	proposals := mocks.NewMockKnowledgeProposalService(t)
	proposals.EXPECT().ListProposalsWithValues(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, params rez.ListProposalsParams) (*ent.ListResult[rez.KnowledgeProposalWithValues], error) {
			require.Equal(t, []kp.Status{kp.StatusPending, kp.StatusRejected}, params.Statuses)
			require.Equal(t, []uuid.UUID{f.checkout.ID, f.checkoutAPI.ID}, params.EntityIDs)
			listed := []*rez.KnowledgeProposalWithValues{f.proposalWithValues()}
			return &ent.ListResult[rez.KnowledgeProposalWithValues]{Data: listed, Total: 1}, nil
		})
	_, api := humatest.New(t)
	handler := newKnowledgeGraphHandler(nil, nil, proposals)
	huma.Register(api, oapi.ListKnowledgeProposals, handler.ListKnowledgeProposals)

	response := api.Get("/knowledge/proposals?status=pending,rejected&entityId=" + f.checkout.ID.String() + "," + f.checkoutAPI.ID.String())

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body oapi.ListKnowledgeProposalsResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body.Body))
	require.Len(t, body.Body.Data, 1)
	values := body.Body.Data[0].Attributes.Values
	require.Len(t, values, 2)
	require.Equal(t, "checkout", values[0].Value)
	require.Equal(t, f.checkout.ID, values[0].Entity.Id)
	require.Equal(t, "Provider B", values[0].Entity.Aliases[0].IntegrationDisplayName)
	require.Equal(t, f.checkoutAPI.ID, values[1].Entity.Id)
	require.Empty(t, body.Body.Data[0].Attributes.Implies)

	t.Run("an unknown status", func(t *testing.T) {
		invalid := api.Get("/knowledge/proposals?status=maybe")

		require.Equal(t, http.StatusUnprocessableEntity, invalid.Code, invalid.Body.String())
	})
}

func TestGetKnowledgeProposal(t *testing.T) {
	f := newKnowledgeFixture()
	proposals := mocks.NewMockKnowledgeProposalService(t)
	proposals.EXPECT().GetProposalWithValues(mock.Anything, f.proposal.ID).Return(f.proposalWithValues(), nil)
	proposals.EXPECT().ListImpliedProposals(mock.Anything, f.proposal).Return([]*ent.KnowledgeProposal{f.rejected}, nil)
	_, api := humatest.New(t)
	handler := newKnowledgeGraphHandler(nil, nil, proposals)
	huma.Register(api, oapi.GetKnowledgeProposal, handler.GetKnowledgeProposal)

	response := api.Get("/knowledge/proposals/" + f.proposal.ID.String())

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body oapi.GetKnowledgeProposalResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body.Body))
	attrs := body.Body.Data.Attributes
	require.Equal(t, "pending", attrs.Status)
	require.Equal(t, "name_similarity", attrs.Origin)
	require.Equal(t, f.proposal.Rationale, attrs.Rationale)

	require.Len(t, attrs.Values, 2)
	require.Equal(t, "checkout", attrs.Values[0].Entity.DisplayName)
	require.Equal(t, "checkout", attrs.Values[0].Entity.Aliases[0].Name)
	require.Equal(t, f.recordedAt, *attrs.Values[0].Entity.Aliases[0].LastRecordedAt)
	require.Equal(t, "Provider A", attrs.Values[1].Entity.Aliases[0].IntegrationDisplayName)

	evidence := attrs.Evidence
	require.Equal(t, f.proposal.Edges.Evidence.ID, evidence.Id)
	require.Equal(t, f.checkout.ID, *evidence.SubjectEntityId)
	require.Equal(t, "Provider B", evidence.IntegrationDisplayName)
	require.Equal(t, "deployment:checkout", evidence.ResourceRef)
	require.Equal(t, "deployment_observed", evidence.Assertion)
	require.Equal(t, f.recordedAt, evidence.EffectiveAt)
	require.Equal(t, "checkout", evidence.DisplayName)

	require.Len(t, attrs.Implies, 1)
	implied := attrs.Implies[0]
	require.Equal(t, f.rejected.ID, implied.Id)
	require.Equal(t, []string{"checkout", "checkout-svc"}, implied.Values)
	require.Equal(t, "rejected", implied.Status)
	require.Equal(t, f.member.ID, implied.RejectedBy.Id)
	require.Equal(t, f.recordedAt, *implied.RejectedAt)
}

func TestConfirmKnowledgeProposalConflict(t *testing.T) {
	f := newKnowledgeFixture()
	proposals := mocks.NewMockKnowledgeProposalService(t)
	refused := errs.Wrap(errs.ErrConflict, errs.CodeConflict, "This proposal is already confirmed.")
	proposals.EXPECT().ConfirmProposal(mock.Anything, f.proposal.ID).Return(nil, refused)
	_, api := humatest.New(t)
	handler := newKnowledgeGraphHandler(nil, nil, proposals)
	huma.Register(api, oapi.ConfirmKnowledgeProposal, handler.ConfirmKnowledgeProposal)

	response := api.Post("/knowledge/proposals/" + f.proposal.ID.String() + "/confirm")

	require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "This proposal is already confirmed.")
}

func TestListKnowledgeGraphEntitiesSearchAndAliases(t *testing.T) {
	f := newKnowledgeFixture()
	aliases := mocks.NewMockKnowledgeEntityAliasService(t)
	aliases.EXPECT().ListEntitiesWithAliases(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, params rez.ListKnowledgeEntitiesParams) (*ent.ListResult[rez.KnowledgeEntityWithAliases], error) {
			require.Equal(t, "check", params.Search)
			require.True(t, params.WithLinkingAttributes)
			listed := []*rez.KnowledgeEntityWithAliases{f.withAliases(f.checkout, "provider-b", "Provider B", "checkout")}
			return &ent.ListResult[rez.KnowledgeEntityWithAliases]{Data: listed, Total: 1}, nil
		})
	_, api := humatest.New(t)
	handler := newKnowledgeGraphHandler(nil, aliases, nil)
	huma.Register(api, oapi.ListKnowledgeGraphEntities, handler.ListKnowledgeGraphEntities)

	response := api.Get("/knowledge/entities?search=check")

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body oapi.ListKnowledgeGraphEntitiesResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body.Body))
	require.Len(t, body.Body.Data, 1)
	attrs := body.Body.Data[0].Attributes
	require.Equal(t, "checkout", attrs.DisplayName)
	require.Equal(t, []oapi.KnowledgeGraphLinkingAttribute{{Attribute: "service.name", Value: "checkout"}}, attrs.LinkingAttributes)
	require.Len(t, attrs.Aliases, 1)
	require.NotEqual(t, uuid.Nil, attrs.Aliases[0].Id)
	require.Equal(t, "entity", attrs.Aliases[0].Attributes.Kind)
	require.Equal(t, "provider-b", attrs.Aliases[0].Attributes.ResourceRef.Provider)
	require.Equal(t, "service:checkout", attrs.Aliases[0].Attributes.ResourceRef.ResourceRef)
	require.Equal(t, "Provider B", attrs.Aliases[0].IntegrationDisplayName)
	require.Equal(t, "provider-b", attrs.Aliases[0].Provider)
	require.Equal(t, "checkout", attrs.Aliases[0].Name)
}
