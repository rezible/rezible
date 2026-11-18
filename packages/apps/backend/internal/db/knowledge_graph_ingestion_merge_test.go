package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	kpd "github.com/rezible/rezible/ent/knowledgeproposaldecision"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/migrate"
	osce "github.com/rezible/rezible/ent/observationscopeentity"
	sae "github.com/rezible/rezible/ent/systemanalysisentity"
	sar "github.com/rezible/rezible/ent/systemanalysisrelationship"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
)

func (s *KnowledgeGraphIngestionServiceSuite) makeProposalService(tdb rez.Database) *KnowledgeProposalService {
	aliases, aliasesErr := NewKnowledgeEntityAliasService(tdb, nil)
	s.Require().NoError(aliasesErr)
	service, serviceErr := NewKnowledgeProposalService(tdb, s.makeService(tdb), aliases, nil)
	s.Require().NoError(serviceErr)
	return service
}

// memberContext acts as a new member of the context's organization.
func (s *KnowledgeGraphIngestionServiceSuite) memberContext(ctx context.Context, tdb rez.Database) (context.Context, *ent.User) {
	createUser := tdb.Client(ctx).User.Create().
		SetEmail(uuid.NewString() + "@example.com").
		SetName("Member")
	user, userErr := createUser.Save(ctx)
	s.Require().NoError(userErr)

	session := &ent.UserAuthSession{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	return execution.NewUserContext(ctx, session), user
}

// pendingCheckoutProposal observes `checkout-api` and then `checkout`, which proposes that they match.
func (s *KnowledgeGraphIngestionServiceSuite) pendingCheckoutProposal(ctx context.Context, tdb rez.Database) *ent.KnowledgeProposal {
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service:checkout-api", "checkout-api"))
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:checkout", "checkout"))
	proposals := s.proposals(ctx, tdb)
	s.Require().Len(proposals, 1)
	return proposals[0]
}

// createProposal records a pending proposal fixture between two names, citing the evidence.
func (s *KnowledgeGraphIngestionServiceSuite) createProposal(ctx context.Context, tdb rez.Database, attribute, valueA, valueB string, evidenceID uuid.UUID) *ent.KnowledgeProposal {
	createProposal := tdb.Client(ctx).KnowledgeProposal.Create().
		SetCategory(kp.CategoryContainer).
		SetKind("service").
		SetAttribute(attribute).
		SetValueA(valueA).
		SetValueB(valueB).
		SetOrigin(kp.OriginNameSimilarity).
		SetRationale("A fixture.").
		SetEvidenceID(evidenceID)
	proposal, createErr := createProposal.Save(ctx)
	s.Require().NoError(createErr)
	return proposal
}

func (s *KnowledgeGraphIngestionServiceSuite) proposalDecisions(ctx context.Context, tdb rez.Database, proposalID uuid.UUID) []*ent.KnowledgeProposalDecision {
	queryDecisions := tdb.Client(ctx).KnowledgeProposalDecision.Query().
		Where(kpd.ProposalID(proposalID)).
		Order(kpd.ByCreatedAt(), kpd.ByID())
	return queryDecisions.AllX(ctx)
}

func (s *KnowledgeGraphIngestionServiceSuite) TestProposalDecisions() {
	s.Run("confirm from pending", func() {
		ctx, tdb := s.SetupTestDatabase()
		proposal := s.pendingCheckoutProposal(ctx, tdb)
		memberCtx, member := s.memberContext(ctx, tdb)
		before := time.Now()

		confirmed, confirmErr := s.makeProposalService(tdb).ConfirmProposal(memberCtx, proposal.ID)

		s.Require().NoError(confirmErr)
		s.Equal(kp.StatusConfirmed, confirmed.Proposal.Status)
		s.Require().Len(confirmed.Proposal.Edges.Decisions, 1)
		decision := confirmed.Proposal.Edges.Decisions[0]
		s.Equal(kpd.StatusConfirmed, decision.Status)
		s.Equal(member.ID, decision.UserID)
		s.Equal(member.ID, decision.Edges.User.ID)
		s.Nil(decision.ImpliedByProposalID)
		s.WithinRange(decision.CreatedAt, before, time.Now())
	})

	s.Run("reject from pending, then confirm from rejected", func() {
		ctx, tdb := s.SetupTestDatabase()
		proposal := s.pendingCheckoutProposal(ctx, tdb)
		memberCtx, member := s.memberContext(ctx, tdb)
		service := s.makeProposalService(tdb)

		rejected, rejectErr := service.RejectProposal(memberCtx, proposal.ID)
		s.Require().NoError(rejectErr)
		s.Equal(kp.StatusRejected, rejected.Proposal.Status)
		s.Equal(2, tdb.Client(ctx).KnowledgeEntity.Query().CountX(ctx))

		confirmed, confirmErr := service.ConfirmProposal(memberCtx, proposal.ID)
		s.Require().NoError(confirmErr)
		s.Equal(kp.StatusConfirmed, confirmed.Proposal.Status)

		decisions := s.proposalDecisions(ctx, tdb, proposal.ID)
		s.Require().Len(decisions, 2)
		s.Equal(kpd.StatusRejected, decisions[0].Status)
		s.Equal(kpd.StatusConfirmed, decisions[1].Status)
		for _, decision := range decisions {
			s.Equal(member.ID, decision.UserID)
		}
	})

	s.Run("a confirmed proposal is final", func() {
		ctx, tdb := s.SetupTestDatabase()
		proposal := s.pendingCheckoutProposal(ctx, tdb)
		memberCtx, _ := s.memberContext(ctx, tdb)
		service := s.makeProposalService(tdb)
		_, confirmErr := service.ConfirmProposal(memberCtx, proposal.ID)
		s.Require().NoError(confirmErr)

		_, rejectErr := service.RejectProposal(memberCtx, proposal.ID)
		_, confirmAgainErr := service.ConfirmProposal(memberCtx, proposal.ID)

		s.ErrorIs(rejectErr, errs.ErrConflict)
		s.ErrorIs(confirmAgainErr, errs.ErrConflict)
		s.Len(s.proposalDecisions(ctx, tdb, proposal.ID), 1)
	})

	s.Run("only a person decides", func() {
		ctx, tdb := s.SetupTestDatabase()
		proposal := s.pendingCheckoutProposal(ctx, tdb)
		service := s.makeProposalService(tdb)

		_, systemErr := service.ConfirmProposal(execution.NewSystemContext(ctx), proposal.ID)
		_, tenantErr := service.RejectProposal(ctx, proposal.ID)

		s.ErrorIs(systemErr, errs.ErrForbidden)
		s.ErrorIs(tenantErr, errs.ErrForbidden)
		s.Equal(kp.StatusPending, tdb.Client(ctx).KnowledgeProposal.GetX(ctx, proposal.ID).Status)
		s.Empty(s.proposalDecisions(ctx, tdb, proposal.ID))
	})
}

func (s *KnowledgeGraphIngestionServiceSuite) TestConfirmMergesEntities() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	graph := newTestGraph(s.Require(), tdb)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	serviceRef := func(provider, resource, name string) rez.KnowledgeEntityRef {
		ref := graph.entity(kne.CategoryContainer, "service", provider, resource)
		ref.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: name}
		return ref
	}

	// `checkout-api` is named by two aliases and observed by an alert first; `checkout` comes later, observed
	// by the same alert.
	checkoutAPI := serviceRef("provider-a", "service:checkout-api", "checkout-api")
	checkoutAPIElsewhere := serviceRef("provider-c", "svc/checkout-api", "checkout-api")
	checkout := serviceRef("provider-b", "service:checkout", "checkout")
	alert := graph.entity(kne.CategorySignal, "alert", "provider-a", "alert:checkout-errors")
	observesCheckoutAPI := graph.relationship(knr.PredicateObserves, alert, checkoutAPI)
	observesCheckout := graph.relationship(knr.PredicateObserves, alert, checkout)
	graph.record(ctx, start,
		graph.observeEntity(checkoutAPI, "checkout-api"),
		graph.observeEntity(alert, "Checkout errors"),
		graph.observeRelationship(observesCheckoutAPI),
	)
	graph.record(ctx, start.Add(time.Minute), graph.observeEntity(checkoutAPIElsewhere, "checkout-api"))
	graph.record(ctx, start.Add(2*time.Minute),
		graph.observeEntity(checkout, "checkout"),
		graph.observeRelationship(observesCheckout),
	)
	survivorID := graph.entityID(ctx, checkoutAPI)
	otherID := graph.entityID(ctx, checkout)
	keptRelationshipID := graph.relationshipID(ctx, observesCheckoutAPI)
	foldedRelationshipID := graph.relationshipID(ctx, observesCheckout)
	proposals := s.proposals(ctx, tdb)
	s.Require().Len(proposals, 1)
	proposal := proposals[0]

	// An analysis holding both relationships and the other service.
	analysis := client.SystemAnalysis.Create().SaveX(ctx)
	for _, relationshipID := range []uuid.UUID{keptRelationshipID, foldedRelationshipID} {
		addRelationship := client.SystemAnalysisRelationship.Create().
			SetAnalysisID(analysis.ID).
			SetKnowledgeRelationshipID(relationshipID)
		addRelationship.ExecX(ctx)
	}
	addEntity := client.SystemAnalysisEntity.Create().
		SetAnalysisID(analysis.ID).
		SetKnowledgeEntityID(otherID)
	addEntity.ExecX(ctx)

	// Discussions on the analysis about the other service and the relationship that folds.
	user := client.User.Query().FirstX(ctx)
	createEntityThread := client.DiscussionThread.Create().
		SetUserID(user.ID).
		SetAnalysisID(analysis.ID).
		SetTargetKnowledgeEntityID(otherID)
	entityThread := createEntityThread.SaveX(ctx)
	createRelationshipThread := client.DiscussionThread.Create().
		SetUserID(user.ID).
		SetAnalysisID(analysis.ID).
		SetTargetKnowledgeRelationshipID(foldedRelationshipID)
	relationshipThread := createRelationshipThread.SaveX(ctx)

	// An observation about both services.
	createDefinition := client.AlertDefinition.Create().
		SetTitle("Checkout errors").
		SetResolutionTimeoutSeconds(0).
		SetMetadataObservedAt(start).
		SetMetadataEventRef("checkout-errors")
	definition := createDefinition.SaveX(ctx)
	createEpisode := client.AlertEpisode.Create().
		SetAlertDefinitionID(definition.ID).
		SetStartedAt(start)
	episode := createEpisode.SaveX(ctx)
	createScope := client.ObservationScope.Create().
		SetAlertEpisodeID(episode.ID).
		SetRevision(1).
		SetBroad(false)
	scope := createScope.SaveX(ctx)
	for _, entityID := range []uuid.UUID{survivorID, otherID} {
		addEntity := client.ObservationScopeEntity.Create().
			SetScopeID(scope.ID).
			SetKnowledgeEntityID(entityID)
		addEntity.ExecX(ctx)
	}

	// A provider read of the other service.
	providerRead := s.startProviderRead(ctx, tdb, otherID)

	memberCtx, member := s.memberContext(ctx, tdb)
	_, confirmErr := s.makeProposalService(tdb).ConfirmProposal(memberCtx, proposal.ID)
	s.Require().NoError(confirmErr)

	s.Run("one service holds both names and every alias", func() {
		queryServices := client.KnowledgeEntity.Query().
			Where(kne.Kind("service"))
		services := queryServices.AllX(ctx)
		s.Require().Len(services, 1)
		s.Equal(survivorID, services[0].ID)
		s.Equal("checkout", services[0].State.DisplayName)
		s.Equal(start.Add(2*time.Minute), services[0].StateEffectiveAt.UTC())
		s.Equal(&survivorID, client.KnowledgeEntity.GetX(ent.IncludeMerged(ctx), otherID).MergedIntoID, "the other is left as a redirect")

		var survivorAliases []knowledgegraph.SubjectAlias
		for _, alias := range client.KnowledgeSubjectAlias.Query().Where(ksa.EntityID(survivorID)).AllX(ctx) {
			survivorAliases = append(survivorAliases, knowledgegraph.SubjectAlias{ID: alias.ID, Authority: alias.Authority, Record: aliasRecord(alias)})
		}
		s.Equal(knowledgegraph.MergeSubjectState(survivorAliases).State, services[0].State, "the survivor's state is a fresh merge of its aliases")

		queryNames := client.KnowledgeEntityLinkingAttribute.Query().
			Where(kela.EntityID(survivorID)).
			Order(kela.ByValue())
		names := queryNames.AllX(ctx)
		s.Require().Len(names, 2)
		s.Equal("checkout", names[0].Value)
		s.Equal("checkout-api", names[1].Value)

		queryAliases := client.KnowledgeSubjectAlias.Query().
			Where(ksa.EntityID(survivorID))
		aliases := queryAliases.AllX(ctx)
		s.Len(aliases, 3)
		for _, alias := range aliases {
			if alias.Provider == "provider-b" {
				s.Equal(&proposal.ID, alias.MergedByProposalID)
			} else {
				s.Nil(alias.MergedByProposalID)
			}
		}
	})

	s.Run("the duplicate relationship is folded with its aliases and analysis row", func() {
		queryRelationships := client.KnowledgeRelationship.Query().
			Where(knr.TargetEntityID(survivorID))
		relationships := queryRelationships.AllX(ctx)
		s.Require().Len(relationships, 1)
		s.Equal(keptRelationshipID, relationships[0].ID)
		folded := client.KnowledgeRelationship.GetX(ent.IncludeMerged(ctx), foldedRelationshipID)
		s.Equal(&keptRelationshipID, folded.MergedIntoID, "the duplicate is left as a redirect")
		s.Equal(otherID, folded.TargetEntityID, "with its endpoints as they were")

		queryRelationshipAliases := client.KnowledgeSubjectAlias.Query().
			Where(ksa.RelationshipID(keptRelationshipID))
		s.Equal(2, queryRelationshipAliases.CountX(ctx))

		queryAnalysisRelationships := client.SystemAnalysisRelationship.Query().
			Where(sar.AnalysisID(analysis.ID))
		analysisRelationships := queryAnalysisRelationships.AllX(ctx)
		s.Require().Len(analysisRelationships, 1)
		s.Equal(keptRelationshipID, analysisRelationships[0].KnowledgeRelationshipID)
	})

	s.Run("references to the other service name the survivor", func() {
		queryScopeEntities := client.ObservationScopeEntity.Query().
			Where(osce.ScopeID(scope.ID))
		s.Equal(survivorID, queryScopeEntities.OnlyX(ctx).KnowledgeEntityID)

		queryAnalysisEntities := client.SystemAnalysisEntity.Query().
			Where(sae.AnalysisID(analysis.ID))
		s.Equal(survivorID, queryAnalysisEntities.OnlyX(ctx).KnowledgeEntityID)

		s.Equal(survivorID, client.AgentProviderRead.GetX(ctx, providerRead.ID).ServiceEntityID)
	})

	s.Run("discussions follow the survivor and the kept relationship", func() {
		s.Equal(&survivorID, client.DiscussionThread.GetX(ctx, entityThread.ID).TargetKnowledgeEntityID)
		s.Equal(&keptRelationshipID, client.DiscussionThread.GetX(ctx, relationshipThread.ID).TargetKnowledgeRelationshipID)
	})

	s.Run("the cited evidence still resolves", func() {
		queryEvidence := client.KnowledgeEvidence.Query().
			Where(ke.ID(proposal.EvidenceID)).
			WithSubjectAlias()
		evidence := queryEvidence.OnlyX(ctx)
		s.Equal(survivorID, *evidence.Edges.SubjectAlias.EntityID)
	})

	s.Run("the moved source says which confirmed match brought it", func() {
		sourcesService, serviceErr := NewKnowledgeEntityAliasService(tdb, nil)
		s.Require().NoError(serviceErr)
		bySurvivor, listErr := sourcesService.ListEntityAliases(ctx, []uuid.UUID{survivorID})
		s.Require().NoError(listErr)

		s.Require().Len(bySurvivor[survivorID], 3)
		for _, source := range bySurvivor[survivorID] {
			merged := source.Alias.Edges.MergedByProposal
			if source.Alias.Provider != "provider-b" {
				s.Nil(merged)
				continue
			}
			s.Equal("checkout", source.Name)
			s.Require().NotNil(merged)
			s.Equal(proposal.ID, merged.ID)
			s.Require().Len(merged.Edges.Decisions, 1)
			s.Equal(member.ID, merged.Edges.Decisions[0].Edges.User.ID)
		}
	})

	s.Run("later events naming either name land on the survivor", func() {
		laterCheckout := serviceRef("provider-d", "service:checkout", "checkout")
		laterCheckoutAPI := serviceRef("provider-e", "service:checkout-api", "checkout-api")

		graph.record(ctx, start.Add(time.Hour), graph.observeEntity(laterCheckout, "checkout"))
		graph.record(ctx, start.Add(time.Hour), graph.observeEntity(laterCheckoutAPI, "checkout-api"))

		s.Equal(survivorID, graph.entityID(ctx, laterCheckout))
		s.Equal(survivorID, graph.entityID(ctx, laterCheckoutAPI))
	})
}

// startProviderRead records a provider read of the service by a running agent turn.
func (s *KnowledgeGraphIngestionServiceSuite) startProviderRead(ctx context.Context, tdb rez.Database, serviceID uuid.UUID) *ent.AgentProviderRead {
	client := tdb.Client(ctx)
	createSession := client.AgentSession.Create().
		SetAgentName("investigation").
		SetInput([]byte(`{}`))
	session := createSession.SaveX(ctx)
	createTurn := client.AgentTurn.Create().
		SetID(uuid.New()).
		SetAgentSessionID(session.ID).
		SetSequence(1).
		SetRiverJobID(1).
		SetStatus(agentturn.StatusRunning)
	turn := createTurn.SaveX(ctx)

	params := rez.StartAgentProviderReadParams{
		AgentTurnID:     turn.ID,
		ToolName:        "search_logs",
		ServiceEntityID: serviceID,
		IntegrationID:   createTestLogIntegration(ctx, s.Require(), tdb),
		Input:           json.RawMessage(`{}`),
	}
	read, _, startErr := NewAgentProviderReadService(tdb, test.NewClock(time.Now())).StartProviderRead(ctx, params)
	s.Require().NoError(startErr)
	return read
}

func (s *KnowledgeGraphIngestionServiceSuite) TestMergeKeepsARelationshipBetweenTheTwo() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	graph := newTestGraph(s.Require(), tdb)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	checkoutAPI := graph.entity(kne.CategoryContainer, "service", "provider-a", "service:checkout-api")
	checkoutAPI.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout-api"}
	checkout := graph.entity(kne.CategoryContainer, "service", "provider-b", "service:checkout")
	checkout.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout"}
	calls := graph.relationship(knr.PredicateCalls, checkoutAPI, checkout)
	graph.record(ctx, start, graph.observeEntity(checkoutAPI, "checkout-api"))
	graph.record(ctx, start.Add(time.Minute),
		graph.observeEntity(checkout, "checkout"),
		graph.observeRelationship(calls),
	)
	survivorID := graph.entityID(ctx, checkoutAPI)
	callsID := graph.relationshipID(ctx, calls)
	proposals := s.proposals(ctx, tdb)
	s.Require().Len(proposals, 1)
	memberCtx, _ := s.memberContext(ctx, tdb)

	_, confirmErr := s.makeProposalService(tdb).ConfirmProposal(memberCtx, proposals[0].ID)

	s.Require().NoError(confirmErr)
	relationship := client.KnowledgeRelationship.GetX(ctx, callsID)
	s.Equal(survivorID, relationship.SourceEntityID)
	s.Equal(survivorID, relationship.TargetEntityID)
	s.Equal(callsID, graph.relationshipID(ctx, calls), "its alias still names it, so its evidence stays citable")
}

// TestMergeFoldsARelationshipIntoASelfLoop merges an entity with a relationship to itself and one from the survivor
// to it: both become one from the survivor to itself, so the later folds into the earlier, and keeps the survivor
// to the merged entity as its endpoints.
func (s *KnowledgeGraphIngestionServiceSuite) TestMergeFoldsARelationshipIntoASelfLoop() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	graph := newTestGraph(s.Require(), tdb)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	checkoutAPI := graph.entity(kne.CategoryContainer, "service", "provider-a", "service:checkout-api")
	checkoutAPI.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout-api"}
	checkout := graph.entity(kne.CategoryContainer, "service", "provider-b", "service:checkout")
	checkout.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout"}
	selfLoop := graph.relationship(knr.PredicateCalls, checkout, checkout)
	fromSurvivor := graph.relationship(knr.PredicateCalls, checkoutAPI, checkout)
	graph.record(ctx, start, graph.observeEntity(checkoutAPI, "checkout-api"))
	graph.record(ctx, start.Add(time.Minute),
		graph.observeEntity(checkout, "checkout"),
		graph.observeRelationship(selfLoop),
	)
	graph.record(ctx, start.Add(2*time.Minute), graph.observeRelationship(fromSurvivor))
	survivorID := graph.entityID(ctx, checkoutAPI)
	otherID := graph.entityID(ctx, checkout)
	selfLoopID := graph.relationshipID(ctx, selfLoop)
	fromSurvivorID := graph.relationshipID(ctx, fromSurvivor)
	proposals := s.proposals(ctx, tdb)
	s.Require().Len(proposals, 1)
	memberCtx, _ := s.memberContext(ctx, tdb)

	_, confirmErr := s.makeProposalService(tdb).ConfirmProposal(memberCtx, proposals[0].ID)

	s.Require().NoError(confirmErr)
	kept := client.KnowledgeRelationship.GetX(ctx, selfLoopID)
	s.Equal(survivorID, kept.SourceEntityID)
	s.Equal(survivorID, kept.TargetEntityID)
	s.Nil(kept.MergedIntoID)
	folded := client.KnowledgeRelationship.GetX(ent.IncludeMerged(ctx), fromSurvivorID)
	s.Equal(&selfLoopID, folded.MergedIntoID)
	s.Equal(survivorID, folded.SourceEntityID, "the folded relationship keeps its endpoints")
	s.Equal(otherID, folded.TargetEntityID, "the folded relationship keeps its endpoints")
}

// TestMergeRecomputesParents merges a container into the container that contains it, and that also contains the
// merged entity's child: the parent claim between them becomes a self-loop, which is never a parent, and the
// child's parent claim folds into the survivor's claim on it.
func (s *KnowledgeGraphIngestionServiceSuite) TestMergeRecomputesParents() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	graph := newTestGraph(s.Require(), tdb)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	checkoutAPI := graph.entity(kne.CategoryContainer, "service", "provider-a", "service:checkout-api")
	checkoutAPI.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout-api"}
	checkout := graph.entity(kne.CategoryContainer, "service", "provider-b", "service:checkout")
	checkout.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout"}
	orders := graph.entity(kne.CategoryComponent, "module", "provider-b", "module:orders")
	graph.record(ctx, start, graph.observeEntity(checkoutAPI, "checkout-api"))
	otherClaim := graph.relationship(knr.PredicateContains, checkout, orders)
	graph.record(ctx, start.Add(time.Minute),
		graph.observeEntity(checkout, "checkout"),
		graph.observeEntity(orders, "orders"),
		graph.observeRelationship(graph.relationship(knr.PredicateContains, checkoutAPI, checkout)),
		graph.observeRelationship(otherClaim),
	)
	// The survivor's own claim on the child is created later, so it ranks after the merged entity's.
	survivorClaim := graph.relationship(knr.PredicateContains, checkoutAPI, orders)
	graph.record(ctx, start.Add(2*time.Minute), graph.observeRelationship(survivorClaim))
	survivorID := graph.entityID(ctx, checkoutAPI)
	otherID := graph.entityID(ctx, checkout)
	ordersID := graph.entityID(ctx, orders)
	otherClaimID := graph.relationshipID(ctx, otherClaim)
	parentClaimOn := func(entityID uuid.UUID) []*ent.KnowledgeRelationship {
		return client.KnowledgeRelationship.Query().
			Where(knr.TargetEntityID(entityID), knr.IsParent(true)).
			AllX(ctx)
	}
	s.Require().Len(parentClaimOn(ordersID), 1)
	s.Require().Equal(otherID, parentClaimOn(ordersID)[0].SourceEntityID, "the merged entity's claim is chosen")
	s.Require().Len(parentClaimOn(otherID), 1)
	s.Require().Equal(survivorID, parentClaimOn(otherID)[0].SourceEntityID)
	assertPlacement(ctx, s.T(), tdb)
	proposals := s.proposals(ctx, tdb)
	s.Require().Len(proposals, 1)
	memberCtx, _ := s.memberContext(ctx, tdb)

	_, confirmErr := s.makeProposalService(tdb).ConfirmProposal(memberCtx, proposals[0].ID)

	s.Require().NoError(confirmErr)
	assertPlacement(ctx, s.T(), tdb)
	s.Empty(parentClaimOn(survivorID), "the self-loop is not a parent")
	ordersClaims := parentClaimOn(ordersID)
	s.Require().Len(ordersClaims, 1)
	s.Equal(survivorID, ordersClaims[0].SourceEntityID)
	s.Equal(graph.relationshipID(ctx, survivorClaim), ordersClaims[0].ID, "the folded claim's flag is on the claim it folded into")
	s.Equal(&ordersClaims[0].ID, client.KnowledgeRelationship.GetX(ent.IncludeMerged(ctx), otherClaimID).MergedIntoID, "the folded claim is left as a redirect")
	queries, queriesErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(queriesErr)
	otherClaims, otherClaimsErr := queries.ListParentClaims(ctx, []uuid.UUID{ordersID})
	s.Require().NoError(otherClaimsErr)
	s.Empty(otherClaims, "the folded claim is not a parent candidate")

	later := start.Add(time.Hour)
	ref := graph.observeRelationship(graph.relationship(knr.PredicateContains, checkoutAPI, checkout))
	ref.EffectiveAt = later
	ingestion := s.makeService(tdb)
	ingester := NewProviderEventIngester(tdb, nil, &evidenceProjector{ingestion, []rez.KnowledgeEvidenceRef{ref}}, ingestion)
	event := s.createEvent(ctx, tdb, "self-contains-restated", later)
	s.Require().NoError(ingester.ProjectNormalizedEvent(ctx, event.ID, false))
	relationshipID := graph.relationshipID(ctx, *ref.Subject.Relationship)
	alias := client.KnowledgeSubjectAlias.Query().Where(ksa.RelationshipID(relationshipID)).OnlyX(ctx)
	s.Require().NotNil(alias.LastObservedAt)
	s.True(later.Equal(*alias.LastObservedAt))
	s.Empty(parentClaimOn(survivorID))
	assertPlacement(ctx, s.T(), tdb)
}

func (s *KnowledgeGraphIngestionServiceSuite) TestMergeFailsWhenARepointedRowRepeatsAUniqueKey() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	proposal := s.pendingCheckoutProposal(ctx, tdb)
	name := projections.LinkingAttributeServiceName
	survivorID := s.entityNamed(ctx, tdb, name, "checkout-api")
	otherID := s.entityNamed(ctx, tdb, name, "checkout")
	// A user's entity is unique and the merge does not delete duplicates of it, so it cannot re-point a user
	// linked to the other service onto the survivor, which another user is linked to.
	for _, entityID := range []uuid.UUID{survivorID, otherID} {
		createUser := client.User.Create().
			SetEmail(uuid.NewString() + "@example.com").
			SetName("Linked").
			SetKnowledgeEntityID(entityID)
		createUser.ExecX(ctx)
	}
	memberCtx, _ := s.memberContext(ctx, tdb)

	_, confirmErr := s.makeProposalService(tdb).ConfirmProposal(memberCtx, proposal.ID)

	s.Require().Error(confirmErr)
	s.True(ent.IsConstraintError(confirmErr), "the unique key refuses the re-pointed row")
	s.True(client.KnowledgeEntity.Query().Where(kne.ID(otherID)).ExistX(ctx))
	s.Equal(otherID, s.entityNamed(ctx, tdb, name, "checkout"), "the other's names stay")
	s.Equal(1, client.KnowledgeSubjectAlias.Query().Where(ksa.EntityID(otherID)).CountX(ctx), "the other's aliases stay")
	s.Equal(kp.StatusPending, client.KnowledgeProposal.GetX(ctx, proposal.ID).Status)
	s.Empty(s.proposalDecisions(ctx, tdb, proposal.ID))
}

func (s *KnowledgeGraphIngestionServiceSuite) TestMergeLocksTheEntitiesHoldingTheNamesNow() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	memberCtx, _ := s.memberContext(ctx, tdb)
	alphaEvent := s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service:alpha", "alpha"))
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:beta", "beta"))
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-c", "service:gamma", "gamma"))
	name := projections.LinkingAttributeServiceName
	alphaID := s.entityNamed(ctx, tdb, name, "alpha")
	betaID := s.entityNamed(ctx, tdb, name, "beta")
	gammaID := s.entityNamed(ctx, tdb, name, "gamma")
	evidenceID := s.eventEvidenceID(ctx, tdb, alphaEvent)
	alphaBeta := s.createProposal(ctx, tdb, name, "alpha", "beta", evidenceID)
	betaGamma := s.createProposal(ctx, tdb, name, "beta", "gamma", evidenceID)

	// `beta` moves to alpha's entity after the caller resolved it to its own.
	_, confirmErr := s.makeProposalService(tdb).ConfirmProposal(memberCtx, alphaBeta.ID)
	s.Require().NoError(confirmErr)
	staleParams := rez.MergeEntitiesParams{
		ProposalID: betaGamma.ID,
		EntityIDs:  [2]uuid.UUID{betaID, gammaID},
	}
	result, mergeErr := s.makeService(tdb).MergeEntities(ctx, staleParams)

	s.Require().NoError(mergeErr)
	s.True(result.Merged)
	s.Equal(alphaID, result.SurvivorID)
	s.Equal(1, client.KnowledgeEntity.Query().Where(kne.Kind("service")).CountX(ctx))
	for _, value := range []string{"alpha", "beta", "gamma"} {
		s.Equal(alphaID, s.entityNamed(ctx, tdb, name, value), value)
	}
}

// TestMergeFlattensRedirects merges gamma into beta, then beta into alpha. An alert observes all three, so each
// merge also folds the merged service's observation into the survivor's. Every redirect names a row that is not
// merged.
func (s *KnowledgeGraphIngestionServiceSuite) TestMergeFlattensRedirects() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	graph := newTestGraph(s.Require(), tdb)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	name := projections.LinkingAttributeServiceName
	alert := graph.entity(kne.CategorySignal, "alert", "provider-a", "alert:errors")
	graph.record(ctx, start, graph.observeEntity(alert, "Errors"))
	var serviceIDs, observesIDs []uuid.UUID
	for i, value := range []string{"alpha", "beta", "gamma"} {
		service := graph.entity(kne.CategoryContainer, "service", "provider-"+value, "service:"+value)
		service.LinkingAttributes = projections.LinkingAttributes{name: value}
		observes := graph.relationship(knr.PredicateObserves, alert, service)
		// Each service is created after the one before, so the earlier one survives each merge.
		graph.record(ctx, start.Add(time.Duration(i+1)*time.Minute),
			graph.observeEntity(service, value),
			graph.observeRelationship(observes),
		)
		serviceIDs = append(serviceIDs, graph.entityID(ctx, service))
		observesIDs = append(observesIDs, graph.relationshipID(ctx, observes))
	}
	evidenceID := graph.latestEvidenceID(ctx, alert.ProviderResourceRef)
	ingestion := s.makeService(tdb)
	merge := func(survivor, other int, valueA, valueB string) {
		proposal := s.createProposal(ctx, tdb, name, valueA, valueB, evidenceID)
		params := rez.MergeEntitiesParams{
			ProposalID: proposal.ID,
			EntityIDs:  [2]uuid.UUID{serviceIDs[survivor], serviceIDs[other]},
		}
		result, mergeErr := ingestion.MergeEntities(ctx, params)
		s.Require().NoError(mergeErr)
		s.Require().Equal(serviceIDs[survivor], result.SurvivorID)
	}

	merge(1, 2, "beta", "gamma")
	merge(0, 1, "alpha", "beta")

	s.Nil(client.KnowledgeEntity.GetX(ctx, serviceIDs[0]).MergedIntoID)
	s.Nil(client.KnowledgeRelationship.GetX(ctx, observesIDs[0]).MergedIntoID)
	withMerged := ent.IncludeMerged(ctx)
	for i, value := range []string{"beta", "gamma"} {
		s.Equal(&serviceIDs[0], client.KnowledgeEntity.GetX(withMerged, serviceIDs[i+1]).MergedIntoID, value)
		s.Equal(&observesIDs[0], client.KnowledgeRelationship.GetX(withMerged, observesIDs[i+1]).MergedIntoID, "observes "+value)
	}
}

func (s *KnowledgeGraphIngestionServiceSuite) TestMergeImpliesOtherProposals() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.makeProposalService(tdb)
	memberCtx, member := s.memberContext(ctx, tdb)
	alphaEvent := s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service:alpha", "alpha"))
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:beta", "beta"))
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-c", "service:gamma", "gamma"))
	evidenceID := s.eventEvidenceID(ctx, tdb, alphaEvent)
	name := projections.LinkingAttributeServiceName
	alphaBeta := s.createProposal(ctx, tdb, name, "alpha", "beta", evidenceID)
	betaGamma := s.createProposal(ctx, tdb, name, "beta", "gamma", evidenceID)
	alphaGamma := s.createProposal(ctx, tdb, name, "alpha", "gamma", evidenceID)
	_, rejectErr := service.RejectProposal(memberCtx, alphaGamma.ID)
	s.Require().NoError(rejectErr)

	impliedByAlphaBeta, impliedErr := service.ListImpliedProposals(ctx, alphaBeta)
	s.Require().NoError(impliedErr)
	s.Empty(impliedByAlphaBeta)

	_, confirmErr := service.ConfirmProposal(memberCtx, alphaBeta.ID)
	s.Require().NoError(confirmErr)
	impliedByBetaGamma, impliedAfterErr := service.ListImpliedProposals(ctx, betaGamma)
	s.Require().NoError(impliedAfterErr)
	s.Require().Len(impliedByBetaGamma, 1)
	s.Equal(alphaGamma.ID, impliedByBetaGamma[0].ID)
	s.Equal(kp.StatusRejected, impliedByBetaGamma[0].Status)

	_, confirmLastErr := service.ConfirmProposal(memberCtx, betaGamma.ID)
	s.Require().NoError(confirmLastErr)

	for _, proposal := range s.proposals(ctx, tdb) {
		s.Equal(kp.StatusConfirmed, proposal.Status, "%s-%s", proposal.ValueA, proposal.ValueB)
	}
	alphaGammaDecisions := s.proposalDecisions(ctx, tdb, alphaGamma.ID)
	s.Require().Len(alphaGammaDecisions, 2)
	s.Equal(kpd.StatusRejected, alphaGammaDecisions[0].Status)
	implied := alphaGammaDecisions[1]
	s.Equal(kpd.StatusConfirmed, implied.Status)
	s.Equal(member.ID, implied.UserID)
	s.Equal(&betaGamma.ID, implied.ImpliedByProposalID)
}

func (s *KnowledgeGraphIngestionServiceSuite) TestMergeRefusesOtherKinds() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	userRef := func(resourceRef string, email string) *rez.KnowledgeEntityRef {
		return &rez.KnowledgeEntityRef{
			Category:            kne.CategoryActor,
			Kind:                "user",
			ProviderResourceRef: s.testProviderRef(resourceRef),
			LinkingAttributes:   testEntityLinkingAttributes{"user.email": email},
		}
	}
	aliceEvent := s.mustObserve(ctx, tdb, userRef("user-1", "alice@example.com"))
	s.mustObserve(ctx, tdb, userRef("user-2", "alice.smith@example.com"))
	proposal := s.createProposal(ctx, tdb, "user.email", "alice.smith@example.com", "alice@example.com", s.eventEvidenceID(ctx, tdb, aliceEvent))
	memberCtx, _ := s.memberContext(ctx, tdb)

	_, confirmErr := s.makeProposalService(tdb).ConfirmProposal(memberCtx, proposal.ID)

	s.ErrorIs(confirmErr, errs.ErrConflict)
	s.Equal(errs.CodeConflict, errs.CodeOf(confirmErr))
	s.Equal(2, client.KnowledgeEntity.Query().CountX(ctx))
	s.Equal(2, client.KnowledgeEntityLinkingAttribute.Query().Where(kela.Attribute("user.email")).CountX(ctx))
	s.Equal(kp.StatusPending, client.KnowledgeProposal.GetX(ctx, proposal.ID).Status)
	s.Empty(s.proposalDecisions(ctx, tdb, proposal.ID))
}

// TestMergeHandlesEveryEntityReference fails when the generated references and the knowledge graph's own columns
// do not cover every foreign key to knowledge entities or relationships, which means the generated code is stale.
// A missed column would keep naming the merged entity's redirect silently.
func TestMergeHandlesEveryEntityReference(t *testing.T) {
	type tableColumn struct{ table, column string }

	// The knowledge graph's own references, which the merge moves in steps of its own.
	handled := mapset.NewSet(
		tableColumn{kela.Table, kela.FieldEntityID},
		tableColumn{ksa.Table, ksa.FieldEntityID},
		tableColumn{ksa.Table, ksa.FieldRelationshipID},
		tableColumn{knr.Table, knr.FieldSourceEntityID},
		tableColumn{knr.Table, knr.FieldTargetEntityID},
		tableColumn{kne.Table, kne.FieldMergedIntoID},
		tableColumn{knr.Table, knr.FieldMergedIntoID},
	)
	for _, reference := range append(ent.KnowledgeEntityReferences, ent.KnowledgeRelationshipReferences...) {
		handled.Add(tableColumn{reference.Table, reference.Column})
	}
	referencing := mapset.NewSet[tableColumn]()
	for _, table := range migrate.Tables {
		for _, foreignKey := range table.ForeignKeys {
			if foreignKey.RefTable != migrate.KnowledgeEntitiesTable && foreignKey.RefTable != migrate.KnowledgeRelationshipsTable {
				continue
			}
			for _, column := range foreignKey.Columns {
				referencing.Add(tableColumn{table.Name, column.Name})
			}
		}
	}
	require.ElementsMatch(t, referencing.ToSlice(), handled.ToSlice(),
		"the generated references and the knowledge graph's own columns are every foreign key to knowledge entities or relationships")
}
