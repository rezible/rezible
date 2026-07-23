package eventprojection

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type ProjectionServiceSuite struct {
	test.Suite
}

func TestProjectionServiceSuite(t *testing.T) {
	suite.Run(t, &ProjectionServiceSuite{Suite: test.NewSuite()})
}

func (s *ProjectionServiceSuite) projectionService() *ProjectionService {
	users, usersErr := db.NewUserService(s.Database(), nil)
	s.Require().NoError(usersErr)

	messageService := mocks.NewMockMessageService(s.T())
	messageService.EXPECT().AddEventHandlers(mock.Anything).Return(nil).Once()

	incidents, incidentsErr := db.NewIncidentService(s.Database(), messageService)
	s.Require().NoError(incidentsErr)

	service, err := NewProjectionService(s.Database(), users, incidents)
	s.Require().NoError(err)
	return service
}

func runProjection(ctx context.Context, service rez.EventProjectionService, event *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	projector, ok := service.GetEventProjectorFunc(event.SubjectKind)
	if !ok {
		return nil, fmt.Errorf("unsupported subject kind %q", event.SubjectKind)
	}
	return projector(ctx, event)
}

func (s *ProjectionServiceSuite) createNormalizedEvent(subjectKind projections.SubjectKind, providerSubjectRef string, kind ne.Kind, occurredAt time.Time, attributes any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encodedAttributes, encodeErr := projections.EncodeAttributes(attributes)
	s.Require().NoError(encodeErr)

	create := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("projection").
		SetProviderEventRef("event-" + uuid.NewString()).
		SetProviderSubjectRef(providerSubjectRef).
		SetKind(kind).
		SetSubjectKind(subjectKind.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt.Add(time.Minute)).
		SetAttributes(encodedAttributes)

	event, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return event
}

func (s *ProjectionServiceSuite) TestGetEventProjectorFuncOnlyReturnsSupportedEventKinds() {
	service := s.projectionService()
	supported := []projections.SubjectKind{
		projections.SubjectKindCodeForge,
		projections.SubjectKindCodeChange,
		projections.SubjectKindSystemComponent,
		projections.SubjectKindSystemRelationship,
		projections.SubjectKindUser,
		projections.SubjectKindIncident,
		projections.SubjectKindIncidentImpact,
		projections.SubjectKindAlertInstance,
	}
	for _, kind := range supported {
		projector, ok := service.GetEventProjectorFunc(kind.String())
		s.True(ok, kind)
		s.NotNil(projector, kind)
	}

	projector, ok := service.GetEventProjectorFunc(projections.SubjectKindChatMessage.String())
	s.False(ok)
	s.Nil(projector)
}

func (s *ProjectionServiceSuite) TestOneEventCanStoreMultipleAssertionsForAnAlias() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()

	attrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: "component:multi",
		Kind:        "service",
		DisplayName: "Multi",
	}
	event := s.createNormalizedEvent(projections.SubjectKindSystemComponent, "component:multi", ne.KindObserved, time.Now(), attrs)

	makeEvidence := func(assertion string) ent.KnowledgeEvidenceRef {
		aliasRef := event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "System component")
		aliasRef.SubjectEntityRef = &ent.KnowledgeEntityRef{
			Kind:        knowledgeEntityKindSystemComponent,
			Reference:   "component:multi",
			DisplayName: "Multi",
		}
		return ent.KnowledgeEvidenceRef{
			Kind:            ke.EvidenceKindObserved,
			Assertion:       assertion,
			EffectiveAt:     event.OccurredAt,
			SubjectAliasRef: aliasRef,
		}
	}

	ingestErr := service.ingestProjectedEvidence(ctx, event, makeEvidence("first"), makeEvidence("second"))
	s.Require().NoError(ingestErr)

	count, countErr := s.Client(ctx).KnowledgeEvidence.Query().Where(ke.EventID(event.ID)).Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(2, count)
}

func (s *ProjectionServiceSuite) TestAliasCannotBeReassignedToAnotherEntity() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()

	subjRef := "component:shared"
	attrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: subjRef,
		Kind:        "service",
		DisplayName: "Shared",
	}
	event := s.createNormalizedEvent(projections.SubjectKindSystemComponent, subjRef, ne.KindObserved, time.Now(), attrs)

	firstAlias := event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "System component")
	firstAlias.SubjectEntityRef = &ent.KnowledgeEntityRef{
		Kind:      knowledgeEntityKindSystemComponent,
		Reference: "component:first",
	}

	firstErr := service.ingestProjectedEvidence(ctx, event, ent.KnowledgeEvidenceRef{
		Kind:            ke.EvidenceKindObserved,
		Assertion:       "first",
		EffectiveAt:     event.OccurredAt,
		SubjectAliasRef: firstAlias,
	})
	s.Require().NoError(firstErr)

	secondAlias := event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "System component")
	secondAlias.SubjectEntityRef = &ent.KnowledgeEntityRef{
		Kind:      knowledgeEntityKindSystemComponent,
		Reference: "component:second",
	}
	secondErr := service.ingestProjectedEvidence(ctx, event, ent.KnowledgeEvidenceRef{
		Kind:            ke.EvidenceKindObserved,
		Assertion:       "second",
		EffectiveAt:     event.OccurredAt,
		SubjectAliasRef: secondAlias,
	})
	s.Require().Error(secondErr)
	s.True(errors.Is(secondErr, rez.ErrConflict))

	queryEntities := s.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ReferenceIn(firstAlias.SubjectEntityRef.Reference, secondAlias.SubjectEntityRef.Reference))
	entityCount, countErr := queryEntities.Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, entityCount)
}
