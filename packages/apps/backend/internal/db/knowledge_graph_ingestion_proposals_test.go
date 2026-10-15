package db

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/projections"
)

// proposalServiceRef is a service as a provider names it, linked by its normalized name.
func (s *KnowledgeGraphIngestionServiceSuite) proposalServiceRef(provider string, resourceRef string, name string) *rez.KnowledgeEntityRef {
	return &rez.KnowledgeEntityRef{
		Category: kne.CategoryContainer,
		Kind:     "service",
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          provider,
			ProviderNamespace: "installation-1",
			ResourceRef:       resourceRef,
		},
		LinkingAttributes: projections.LinkingAttributes{
			projections.LinkingAttributeServiceName: projections.NormalizeServiceName(name),
		},
	}
}

// observe ingests one event observing the entity and returns the event.
func (s *KnowledgeGraphIngestionServiceSuite) observe(ctx context.Context, tdb rez.Database, ref *rez.KnowledgeEntityRef) (*ent.NormalizedEvent, error) {
	now := time.Now().UTC()
	event := s.createEvent(ctx, tdb, "event:"+uuid.NewString(), now)
	evidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "entity_observed",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: ref.ProviderResourceRef.ResourceRef},
		Subject:      rez.KnowledgeSubjectRef{Entity: ref},
	}
	return event, s.makeService(tdb).IngestEvidence(ctx, event, evidence)
}

func (s *KnowledgeGraphIngestionServiceSuite) mustObserve(ctx context.Context, tdb rez.Database, ref *rez.KnowledgeEntityRef) *ent.NormalizedEvent {
	event, ingestErr := s.observe(ctx, tdb, ref)
	s.Require().NoError(ingestErr)
	return event
}

func (s *KnowledgeGraphIngestionServiceSuite) proposals(ctx context.Context, tdb rez.Database) []*ent.KnowledgeProposal {
	queryProposals := tdb.Client(ctx).KnowledgeProposal.Query().
		Order(kp.ByNameA(), kp.ByNameB())
	return queryProposals.AllX(ctx)
}

func (s *KnowledgeGraphIngestionServiceSuite) eventEvidenceID(ctx context.Context, tdb rez.Database, event *ent.NormalizedEvent) uuid.UUID {
	evidenceQuery := tdb.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.EventID(event.ID))
	return evidenceQuery.OnlyIDX(ctx)
}

func (s *KnowledgeGraphIngestionServiceSuite) entityNamed(ctx context.Context, tdb rez.Database, attribute string, value string) uuid.UUID {
	valueQuery := tdb.Client(ctx).KnowledgeEntityLinkingAttribute.Query().
		Where(kela.Attribute(attribute), kela.Value(value))
	return valueQuery.OnlyX(ctx).EntityID
}

func (s *KnowledgeGraphIngestionServiceSuite) TestSimilarNameBecomesProposal() {
	orders := []struct {
		name          string
		first, second *rez.KnowledgeEntityRef
	}{
		{
			name:   "checkout-api first",
			first:  s.proposalServiceRef("provider-a", "service:checkout-api", "checkout-api"),
			second: s.proposalServiceRef("provider-b", "service:checkout", "checkout"),
		},
		{
			name:   "checkout first",
			first:  s.proposalServiceRef("provider-b", "service:checkout", "checkout"),
			second: s.proposalServiceRef("provider-a", "service:checkout-api", "checkout-api"),
		},
	}
	for _, order := range orders {
		s.Run(order.name, func() {
			ctx, tdb := s.SetupTestDatabase()
			s.mustObserve(ctx, tdb, order.first)

			secondEvent := s.mustObserve(ctx, tdb, order.second)

			s.Equal(2, tdb.Client(ctx).KnowledgeEntity.Query().CountX(ctx))
			proposals := s.proposals(ctx, tdb)
			s.Require().Len(proposals, 1)
			proposal := proposals[0]
			s.Equal(kp.StatusPending, proposal.Status)
			s.Equal(kp.OriginNameSimilarity, proposal.Origin)
			s.Equal(kp.CategoryContainer, proposal.Category)
			s.Equal("service", proposal.Kind)
			s.Equal(projections.LinkingAttributeServiceName, proposal.Attribute)
			s.Equal("checkout", proposal.NameA)
			s.Equal("checkout-api", proposal.NameB)
			s.Contains(proposal.Rationale, "differ only by the generic term `api`")
			s.Equal(s.eventEvidenceID(ctx, tdb, secondEvent), proposal.EvidenceID)

			// The second event again, and a third provider naming an existing name, add nothing.
			ingestAgainErr := s.makeService(tdb).IngestEvidence(ctx, secondEvent, rez.KnowledgeEvidenceRef{
				Kind:        ke.KindObserved,
				Assertion:   "entity_observed",
				EffectiveAt: time.Now().UTC(),
				Subject:     rez.KnowledgeSubjectRef{Entity: order.second},
			})
			s.Require().NoError(ingestAgainErr)
			s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-c", "service:checkout", "checkout"))
			s.Len(s.proposals(ctx, tdb), 1)
		})
	}

	s.Run("a different process of the same product is not proposed", func() {
		ctx, tdb := s.SetupTestDatabase()
		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service:payments-api", "payments-api"))

		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:payments-worker", "payments-worker"))

		s.Empty(s.proposals(ctx, tdb))
	})

	s.Run("a name whose stem matches nothing is not proposed", func() {
		ctx, tdb := s.SetupTestDatabase()
		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service:checkout-api", "checkout-api"))

		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:inventory-api", "inventory-api"))

		s.Empty(s.proposals(ctx, tdb))
	})

	s.Run("an entity holding both names is not proposed against itself", func() {
		ctx, tdb := s.SetupTestDatabase()
		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service-1", "checkout"))

		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service-1", "checkout-api"))

		s.Equal(1, tdb.Client(ctx).KnowledgeEntity.Query().CountX(ctx))
		s.Equal(2, tdb.Client(ctx).KnowledgeEntityLinkingAttribute.Query().CountX(ctx))
		s.Empty(s.proposals(ctx, tdb))
	})

	s.Run("a rejected pair is not proposed again", func() {
		ctx, tdb := s.SetupTestDatabase()
		firstEvent := s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service:checkout-api", "checkout-api"))
		createRejected := tdb.Client(ctx).KnowledgeProposal.Create().
			SetCategory(kp.CategoryContainer).
			SetKind("service").
			SetAttribute(projections.LinkingAttributeServiceName).
			SetNameA("checkout").
			SetNameB("checkout-api").
			SetStatus(kp.StatusRejected).
			SetOrigin(kp.OriginNameSimilarity).
			SetRationale("Rejected earlier.").
			SetEvidenceID(s.eventEvidenceID(ctx, tdb, firstEvent))
		createRejected.ExecX(ctx)

		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:checkout", "checkout"))

		proposals := s.proposals(ctx, tdb)
		s.Require().Len(proposals, 1)
		s.Equal(kp.StatusRejected, proposals[0].Status)
	})

	s.Run("a repository named like another is not proposed", func() {
		ctx, tdb := s.SetupTestDatabase()
		repositoryRef := func(fullName string) *rez.KnowledgeEntityRef {
			return &rez.KnowledgeEntityRef{
				Category:            kne.CategoryCode,
				Kind:                "repository",
				ProviderResourceRef: s.testProviderRef("repository:" + fullName),
				LinkingAttributes: projections.LinkingAttributes{
					projections.LinkingAttributeRepositoryFullName: fullName,
				},
			}
		}
		s.mustObserve(ctx, tdb, repositoryRef("acme/checkout"))

		s.mustObserve(ctx, tdb, repositoryRef("acme/checkout-api"))

		s.Empty(s.proposals(ctx, tdb))
	})

	s.Run("similar names ingested concurrently are proposed once", func() {
		for range 5 {
			ctx, tdb := s.SetupTestDatabase()
			refs := []*rez.KnowledgeEntityRef{
				s.proposalServiceRef("provider-a", "service:checkout-api", "checkout-api"),
				s.proposalServiceRef("provider-b", "service:checkout", "checkout"),
			}
			ingestErrs := make([]error, len(refs))

			var wg sync.WaitGroup
			for i, ref := range refs {
				wg.Go(func() {
					_, ingestErrs[i] = s.observe(ctx, tdb, ref)
				})
			}
			wg.Wait()

			for _, ingestErr := range ingestErrs {
				s.Require().NoError(ingestErr)
			}
			s.Len(s.proposals(ctx, tdb), 1)
		}
	})
}

func (s *KnowledgeGraphIngestionServiceSuite) TestIdentityConflictBecomesProposal() {
	s.Run("an alias naming another service's name stays on its entity", func() {
		ctx, tdb := s.SetupTestDatabase()
		client := tdb.Client(ctx)
		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service-1", "checkout"))
		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:payments", "payments"))
		checkoutID := s.entityNamed(ctx, tdb, projections.LinkingAttributeServiceName, "checkout")
		paymentsID := s.entityNamed(ctx, tdb, projections.LinkingAttributeServiceName, "payments")

		renamed := s.proposalServiceRef("provider-a", "service-1", "payments")
		conflictEvent := s.mustObserve(ctx, tdb, renamed)

		conflictEvidenceQuery := client.KnowledgeEvidence.Query().
			Where(ke.EventID(conflictEvent.ID)).
			WithSubjectAlias()
		conflictEvidence := conflictEvidenceQuery.OnlyX(ctx)
		s.Equal(checkoutID, *conflictEvidence.Edges.SubjectAlias.EntityID)
		checkoutValuesQuery := client.KnowledgeEntityLinkingAttribute.Query().
			Where(kela.EntityID(checkoutID))
		s.Equal(1, checkoutValuesQuery.CountX(ctx))
		s.Equal(paymentsID, s.entityNamed(ctx, tdb, projections.LinkingAttributeServiceName, "payments"))

		proposals := s.proposals(ctx, tdb)
		s.Require().Len(proposals, 1)
		s.Equal(kp.StatusPending, proposals[0].Status)
		s.Equal(kp.OriginIdentityConflict, proposals[0].Origin)
		s.Equal("checkout", proposals[0].NameA)
		s.Equal("payments", proposals[0].NameB)
		s.Equal(conflictEvidence.ID, proposals[0].EvidenceID)

		againErr := s.makeService(tdb).IngestEvidence(ctx, conflictEvent, rez.KnowledgeEvidenceRef{
			Kind:        ke.KindObserved,
			Assertion:   "entity_observed",
			EffectiveAt: time.Now().UTC(),
			Subject:     rez.KnowledgeSubjectRef{Entity: renamed},
		})
		s.Require().NoError(againErr)
		s.Len(s.proposals(ctx, tdb), 1)
		aliasEvidenceQuery := client.KnowledgeEvidence.Query().
			Where(ke.HasSubjectAliasWith(ksa.ProviderResourceRef("service-1")))
		s.Equal(2, aliasEvidenceQuery.CountX(ctx))
	})

	s.Run("an alias whose entity holds no name still fails", func() {
		ctx, tdb := s.SetupTestDatabase()
		unnamed := s.proposalServiceRef("provider-a", "service-1", "checkout")
		unnamed.LinkingAttributes = nil
		s.mustObserve(ctx, tdb, unnamed)
		s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:payments", "payments"))

		_, conflictErr := s.observe(ctx, tdb, s.proposalServiceRef("provider-a", "service-1", "payments"))

		s.ErrorIs(conflictErr, errs.ErrConflict)
		s.Empty(s.proposals(ctx, tdb))
	})

	s.Run("a conflict on a user still fails", func() {
		ctx, tdb := s.SetupTestDatabase()
		userRef := func(resourceRef string, email string) *rez.KnowledgeEntityRef {
			return &rez.KnowledgeEntityRef{
				Category:            kne.CategoryActor,
				Kind:                "user",
				ProviderResourceRef: s.testProviderRef(resourceRef),
				LinkingAttributes:   testEntityLinkingAttributes{"user.email": email},
			}
		}
		s.mustObserve(ctx, tdb, userRef("user-1", "alice@example.com"))
		s.mustObserve(ctx, tdb, userRef("user-2", "bob@example.com"))

		_, conflictErr := s.observe(ctx, tdb, userRef("user-1", "bob@example.com"))

		s.ErrorIs(conflictErr, errs.ErrConflict)
		s.Empty(s.proposals(ctx, tdb))
	})
}

func (s *KnowledgeGraphIngestionServiceSuite) TestListProposals() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	proposalService, serviceErr := NewKnowledgeProposalService(tdb)
	s.Require().NoError(serviceErr)
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service:checkout-api", "checkout-api"))
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:checkout", "checkout"))
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service:payments-api", "payments-api"))
	s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:payments-svc", "payments-svc"))
	rejectPayments := client.KnowledgeProposal.Update().
		Where(kp.NameA("payments-api")).
		SetStatus(kp.StatusRejected)
	rejectPayments.ExecX(ctx)
	checkoutID := s.entityNamed(ctx, tdb, projections.LinkingAttributeServiceName, "checkout")
	checkoutAPIID := s.entityNamed(ctx, tdb, projections.LinkingAttributeServiceName, "checkout-api")
	list := func(params rez.ListProposalsParams) *ent.ListResult[ent.KnowledgeProposal] {
		result, listErr := proposalService.ListProposals(ctx, params)
		s.Require().NoError(listErr)
		return result
	}

	s.Run("by status", func() {
		pending := list(rez.ListProposalsParams{Statuses: []kp.Status{kp.StatusPending}})
		s.Require().Len(pending.Data, 1)
		s.Equal("checkout", pending.Data[0].NameA)

		rejected := list(rez.ListProposalsParams{Statuses: []kp.Status{kp.StatusRejected}})
		s.Require().Len(rejected.Data, 1)
		s.Equal("payments-api", rejected.Data[0].NameA)
	})

	s.Run("by entity, with evidence", func() {
		forCheckout := list(rez.ListProposalsParams{EntityIDs: []uuid.UUID{checkoutID}})

		s.Require().Len(forCheckout.Data, 1)
		proposal := forCheckout.Data[0]
		s.Equal("checkout", proposal.NameA)
		evidence, evidenceErr := proposal.Edges.EvidenceOrErr()
		s.Require().NoError(evidenceErr)
		alias, aliasErr := evidence.Edges.SubjectAliasOrErr()
		s.Require().NoError(aliasErr)
		s.Equal("provider-b", alias.Provider)
	})

	s.Run("paged", func() {
		firstPage := list(rez.ListProposalsParams{ListParams: ent.ListParams{PageSize: 1}})

		s.Len(firstPage.Data, 1)
		s.Equal(2, firstPage.Total)
	})

	s.Run("a value moved to another entity is listed for the entity holding it", func() {
		moveCheckout := client.KnowledgeEntityLinkingAttribute.Delete().
			Where(kela.Value("checkout"))
		moveCheckout.ExecX(ctx)
		addCheckout := client.KnowledgeEntityLinkingAttribute.Create().
			SetEntityID(checkoutAPIID).
			SetAttribute(projections.LinkingAttributeServiceName).
			SetValue("checkout")
		addCheckout.ExecX(ctx)

		forSurvivor := list(rez.ListProposalsParams{EntityIDs: []uuid.UUID{checkoutAPIID}})
		forMergedAway := list(rez.ListProposalsParams{EntityIDs: []uuid.UUID{checkoutID}})

		s.Require().Len(forSurvivor.Data, 1)
		s.Equal("checkout", forSurvivor.Data[0].NameA)
		s.Empty(forMergedAway.Data)
	})
}
