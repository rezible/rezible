package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/review"
	"github.com/rezible/rezible/ent/userauthsession"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type RetrospectiveServiceSuite struct {
	test.Suite
}

func TestRetrospectiveServiceSuite(t *testing.T) {
	suite.Run(t, &RetrospectiveServiceSuite{Suite: test.NewSuite()})
}

func (s *RetrospectiveServiceSuite) createIncident(ctx context.Context, client *ent.Client) *ent.Incident {
	severity, err := client.IncidentSeverity.Create().
		SetName("SEV-1 " + uuid.NewString()).
		SetRank(1).
		SetDescription("Critical").
		Save(ctx)
	s.Require().NoError(err)

	incidentType, err := client.IncidentType.Create().
		SetName("Customer Impact " + uuid.NewString()).
		Save(ctx)
	s.Require().NoError(err)

	inc, err := client.Incident.Create().
		SetSlug("incident-" + uuid.NewString()).
		SetTitle("API outage").
		SetSeverityID(severity.ID).
		SetTypeID(incidentType.ID).
		SetResponseState(incident.ResponseStateResolved).
		SetResolvedAt(time.Now()).
		Save(ctx)
	s.Require().NoError(err)
	return inc
}

func (s *RetrospectiveServiceSuite) TestCreateFullRetrospective() {
	ctx, tdb := s.SetupTestDatabase()
	svc := &RetrospectiveService{db: tdb}

	client := tdb.Client(ctx)
	inc := s.createIncident(ctx, client)

	retro, err := svc.CreateForIncident(ctx, inc.ID)
	s.Require().NoError(err)
	s.NotNil(retro)
	s.NotEqual(uuid.Nil, retro.DocumentID)
	s.NotEqual(uuid.Nil, retro.SystemAnalysisID)
	_, analysisErr := client.SystemAnalysis.Get(ctx, retro.SystemAnalysisID)
	s.Require().NoError(analysisErr)
}

func (s *RetrospectiveServiceSuite) TestRejectUnresolvedIncident() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	inc := s.createIncident(ctx, client)
	unresolve := client.Incident.UpdateOneID(inc.ID).
		SetResponseState(incident.ResponseStateStarted).
		ClearResolvedAt()
	s.Require().NoError(unresolve.Exec(ctx))
	svc := &RetrospectiveService{db: tdb}
	retro, createErr := svc.CreateForIncident(ctx, inc.ID)
	s.ErrorIs(createErr, errs.ErrConflict)
	s.Nil(retro)
	s.Zero(client.Retrospective.Query().CountX(ctx))
	s.Zero(client.Document.Query().CountX(ctx))
	s.Zero(client.SystemAnalysis.Query().CountX(ctx))
}

func (s *RetrospectiveServiceSuite) TestConcurrentCreationReusesWorkspace() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	inc := s.createIncident(ctx, client)
	svc := &RetrospectiveService{db: tdb}
	const callers = 8
	results := make([]*ent.Retrospective, callers)
	failures := make([]error, callers)
	var callersDone sync.WaitGroup
	start := make(chan struct{})
	for i := range callers {
		callersDone.Go(func() {
			<-start
			results[i], failures[i] = svc.CreateForIncident(ctx, inc.ID)
		})
	}
	close(start)
	callersDone.Wait()
	for i := range callers {
		s.Require().NoError(failures[i])
		s.Require().NotNil(results[i])
		s.Equal(results[0].ID, results[i].ID)
		s.Equal(retrospective.StateDraft, results[i].State)
	}
	s.Equal(1, client.Retrospective.Query().CountX(ctx))
	s.Equal(1, client.Document.Query().CountX(ctx))
	s.Equal(1, client.SystemAnalysis.Query().CountX(ctx))
}

func (s *RetrospectiveServiceSuite) TestCreationRollsBackWithTransaction() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	inc := s.createIncident(ctx, client)
	svc := &RetrospectiveService{db: tdb}
	abort := fmt.Errorf("abort incident update")
	txErr := tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		retro, createErr := svc.CreateForIncident(txCtx, inc.ID)
		s.Require().NoError(createErr)
		s.Require().NotNil(retro)
		return abort
	})
	s.ErrorIs(txErr, abort)
	s.Zero(client.Retrospective.Query().CountX(ctx))
	s.Zero(client.Document.Query().CountX(ctx))
	s.Zero(client.SystemAnalysis.Query().CountX(ctx))
}

// adminContext acts as an admin who is not the retrospective's owner.
func (s *RetrospectiveServiceSuite) adminContext(ctx context.Context, userID uuid.UUID) context.Context {
	tenant, _ := execution.GetContext(ctx).TenantID()
	return execution.NewUserContext(ctx, &ent.UserAuthSession{UserID: userID, TenantID: tenant, Role: userauthsession.RoleAdmin})
}

func (s *RetrospectiveServiceSuite) TestReviewTransitions() {
	cases := []struct {
		name    string
		run     func(context.Context, *RetrospectiveService, *ent.Retrospective, uuid.UUID, uuid.UUID) error
		code    errs.Code
		state   retrospective.State
		message string
	}{
		{"request review", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			_, actionErr := svc.RequestReview(ctx, r.ID, []uuid.UUID{reviewer, reviewer})
			return actionErr
		}, "", retrospective.StateInReview, ""},
		{"close without review", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			_, actionErr := svc.CloseReview(ctx, r.ID)
			return actionErr
		}, "", retrospective.StateClosed, ""},
		{"skip needs reason", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			_, actionErr := svc.SkipReview(ctx, r.ID, " ")
			return actionErr
		}, errs.CodeInvalidInput, retrospective.StateDraft, ""},
		{"last reviewer returns to draft", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			if _, requestErr := svc.RequestReview(ctx, r.ID, []uuid.UUID{reviewer}); requestErr != nil {
				return requestErr
			}
			_, actionErr := svc.RemoveReviewer(ctx, r.ID, reviewer)
			return actionErr
		}, "", retrospective.StateDraft, ""},
		{"waiting blocks close", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			if _, requestErr := svc.RequestReview(ctx, r.ID, []uuid.UUID{reviewer}); requestErr != nil {
				return requestErr
			}
			_, actionErr := svc.CloseReview(ctx, r.ID)
			return actionErr
		}, errs.CodeConflict, retrospective.StateInReview, "Every reviewer must approve before closing. Waiting on 1 reviewer."},
		{"approved permits close", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			if _, requestErr := svc.RequestReview(ctx, r.ID, []uuid.UUID{reviewer}); requestErr != nil {
				return requestErr
			}
			tenant, _ := execution.GetContext(ctx).TenantID()
			reviewerCtx := execution.NewUserContext(ctx, &ent.UserAuthSession{UserID: reviewer, TenantID: tenant})
			decision := rez.SetReviewDecisionParams{ReviewerID: reviewer, Decision: review.StateApproved}
			if _, decideErr := svc.SetReviewDecision(reviewerCtx, r.ID, decision); decideErr != nil {
				return decideErr
			}
			_, actionErr := svc.CloseReview(ctx, r.ID)
			return actionErr
		}, "", retrospective.StateClosed, ""},
		{"non-owner rejected", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			tenant, _ := execution.GetContext(ctx).TenantID()
			otherCtx := execution.NewUserContext(ctx, &ent.UserAuthSession{UserID: reviewer, TenantID: tenant})
			_, actionErr := svc.CloseReview(otherCtx, r.ID)
			return actionErr
		}, errs.CodeForbidden, retrospective.StateDraft, ""},
		{"unowned only takes ownership for self", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			if clearErr := svc.db.Client(ctx).Retrospective.UpdateOneID(r.ID).ClearOwner().Exec(ctx); clearErr != nil {
				return clearErr
			}
			tenant, _ := execution.GetContext(ctx).TenantID()
			otherCtx := execution.NewUserContext(ctx, &ent.UserAuthSession{UserID: reviewer, TenantID: tenant})
			if _, assignErr := svc.SetOwner(otherCtx, r.ID, owner); errs.CodeOf(assignErr) != errs.CodeForbidden {
				return fmt.Errorf("assigning another member: %w", assignErr)
			}
			_, actionErr := svc.SetOwner(otherCtx, r.ID, reviewer)
			return actionErr
		}, "", retrospective.StateDraft, ""},
		{"admin closes another member's review", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			_, actionErr := svc.CloseReview(s.adminContext(ctx, reviewer), r.ID)
			return actionErr
		}, "", retrospective.StateClosed, ""},
		{"admin skips another member's review", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			_, actionErr := svc.SkipReview(s.adminContext(ctx, reviewer), r.ID, "Duplicate")
			return actionErr
		}, "", retrospective.StateClosed, ""},
		{"admin hands over another member's review", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			createUser := svc.db.Client(ctx).User.Create().
				SetName("New owner").
				SetEmail(uuid.NewString() + "@example.com")
			newOwner, createErr := createUser.Save(ctx)
			if createErr != nil {
				return createErr
			}
			updated, actionErr := svc.SetOwner(s.adminContext(ctx, reviewer), r.ID, newOwner.ID)
			if actionErr != nil {
				return actionErr
			}
			if *updated.OwnerID != newOwner.ID {
				return fmt.Errorf("owner is %s, want %s", *updated.OwnerID, newOwner.ID)
			}
			return nil
		}, "", retrospective.StateDraft, ""},
		{"admin assigns an owner before closing an unowned review", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			if clearErr := svc.db.Client(ctx).Retrospective.UpdateOneID(r.ID).ClearOwner().Exec(ctx); clearErr != nil {
				return clearErr
			}
			adminCtx := s.adminContext(ctx, reviewer)
			if _, closeErr := svc.CloseReview(adminCtx, r.ID); errs.CodeOf(closeErr) != errs.CodeConflict {
				return fmt.Errorf("closing unowned review: %w", closeErr)
			}
			if _, skipErr := svc.SkipReview(adminCtx, r.ID, "Duplicate"); errs.CodeOf(skipErr) != errs.CodeConflict {
				return fmt.Errorf("skipping unowned review: %w", skipErr)
			}
			_, actionErr := svc.SetOwner(adminCtx, r.ID, owner)
			return actionErr
		}, "", retrospective.StateDraft, ""},
		{"decision only for self", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			if _, requestErr := svc.RequestReview(ctx, r.ID, []uuid.UUID{reviewer}); requestErr != nil {
				return requestErr
			}
			decision := rez.SetReviewDecisionParams{ReviewerID: reviewer, Decision: review.StateApproved}
			_, actionErr := svc.SetReviewDecision(ctx, r.ID, decision)
			return actionErr
		}, errs.CodeForbidden, retrospective.StateInReview, ""},
		{"request needs owner", func(ctx context.Context, svc *RetrospectiveService, r *ent.Retrospective, owner, reviewer uuid.UUID) error {
			update := svc.db.Client(ctx).Retrospective.UpdateOneID(r.ID).ClearOwner()
			if updateErr := update.Exec(ctx); updateErr != nil {
				return updateErr
			}
			_, actionErr := svc.RequestReview(ctx, r.ID, []uuid.UUID{reviewer})
			return actionErr
		}, errs.CodeConflict, retrospective.StateDraft, ""},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			ctx, tdb := s.SetupTestDatabase()
			client := tdb.Client(ctx)
			owner, ownerErr := client.User.Query().First(ctx)
			s.Require().NoError(ownerErr)
			createReviewer := client.User.Create().
				SetName("Reviewer").
				SetEmail(uuid.NewString() + "@example.com")
			reviewer, reviewerErr := createReviewer.Save(ctx)
			s.Require().NoError(reviewerErr)
			ctx = execution.NewUserContext(ctx, &ent.UserAuthSession{UserID: owner.ID, TenantID: owner.TenantID})
			jobService := mocks.NewMockJobService(s.T())
			jobService.EXPECT().Insert(mock.Anything, mock.Anything, mock.Anything).Return(&rivertype.JobInsertResult{}, nil).Maybe()
			svc, svcErr := NewRetrospectiveService(tdb, jobService)
			s.Require().NoError(svcErr)
			inc := s.createIncident(ctx, client)
			setDeclarer := client.Incident.UpdateOneID(inc.ID).SetDeclaredByID(owner.ID)
			s.Require().NoError(setDeclarer.Exec(ctx))
			retro, createErr := svc.CreateForIncident(ctx, inc.ID)
			s.Require().NoError(createErr)
			s.Require().Equal(owner.ID, *retro.OwnerID)
			actionErr := tc.run(ctx, svc, retro, owner.ID, reviewer.ID)
			if tc.code == "" {
				s.Require().NoError(actionErr)
			} else {
				s.Equal(tc.code, errs.CodeOf(actionErr))
			}
			if tc.message != "" {
				s.Equal(tc.message, errs.PublicMessage(actionErr))
			}
			loaded, getErr := svc.GetRetrospective(ctx, retrospective.ID(retro.ID))
			s.Require().NoError(getErr)
			s.Equal(tc.state, loaded.State)
			if loaded.State == retrospective.StateClosed {
				s.NotNil(loaded.ClosedAt)
				s.NotNil(loaded.ClosedByID)
				_, repeatErr := svc.SkipReview(ctx, retro.ID, "again")
				s.Equal(errs.CodeConflict, errs.CodeOf(repeatErr))
			}
		})
	}
}

func (s *RetrospectiveServiceSuite) TestReviewerMessages() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	owner, ownerErr := client.User.Query().First(ctx)
	s.Require().NoError(ownerErr)
	createReviewer := client.User.Create().
		SetName("Reviewer").
		SetEmail(uuid.NewString() + "@example.com")
	reviewer, reviewerErr := createReviewer.Save(ctx)
	s.Require().NoError(reviewerErr)
	ctx = execution.NewUserContext(ctx, &ent.UserAuthSession{UserID: owner.ID, TenantID: owner.TenantID})
	jobService := mocks.NewMockJobService(s.T())
	jobService.EXPECT().Insert(mock.Anything, mock.Anything, mock.Anything).Return(&rivertype.JobInsertResult{}, nil)
	installations := mocks.NewMockIntegrationService(s.T())
	svc, svcErr := NewRetrospectiveService(tdb, jobService)
	s.Require().NoError(svcErr)
	inc := s.createIncident(ctx, client)
	setDeclarer := client.Incident.UpdateOneID(inc.ID).SetDeclaredByID(owner.ID)
	s.Require().NoError(setDeclarer.Exec(ctx))
	retro, createErr := svc.CreateForIncident(ctx, inc.ID)
	s.Require().NoError(createErr)
	requested, requestErr := svc.RequestReview(ctx, retro.ID, []uuid.UUID{reviewer.ID})
	s.Require().NoError(requestErr)
	row := requested.Edges.Reviews[0]
	sender := &reviewMessageSender{}
	installation := &ent.Integration{ID: uuid.New()}
	fallbackInstallation := &ent.Integration{ID: uuid.New()}
	fallback := &reviewMessageSender{}
	installations.EXPECT().ListInstallations(mock.Anything).Return([]*ent.Integration{installation, fallbackInstallation}, nil)
	installations.EXPECT().AsInstalledIntegration(installation).Return(sender, nil)
	installations.EXPECT().AsInstalledIntegration(fallbackInstallation).Return(fallback, nil)
	worker := &sendReviewRequestMessageWorker{db: tdb, cfg: s.Config(), integrations: installations}
	job := &river.Job[jobs.SendReviewRequestMessage]{JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 3}, Args: jobs.SendReviewRequestMessage{ReviewID: row.ID}}
	s.Require().NoError(worker.Work(ctx, job))
	s.Equal(reviewer.ID, sender.recipient)
	s.Contains(sender.message.Text, owner.Name)
	s.Contains(sender.message.LinkURL, inc.Slug)
	sender.sendErr = errors.New("chat unavailable")
	s.Require().NoError(worker.Work(ctx, job))
	s.Equal(reviewer.ID, fallback.recipient)
	fallback.sendErr = errors.New("other chat unavailable")
	s.Require().Error(worker.Work(ctx, job))
	stored, storedErr := client.Review.Get(ctx, row.ID)
	s.Require().NoError(storedErr)
	s.Nil(stored.MessageFailedAt)
	job.Attempt = 3
	s.Require().Error(worker.Work(ctx, job))
	stored, storedErr = client.Review.Get(ctx, row.ID)
	s.Require().NoError(storedErr)
	s.NotNil(stored.MessageFailedAt)
	archive := client.Incident.UpdateOneID(inc.ID).SetArchiveTime(time.Now())
	s.Require().NoError(archive.Exec(ctx))
	s.Require().NoError(worker.Work(ctx, job), "an archived incident ends the job before the failing senders")
}

type reviewMessageSender struct {
	rez.InstalledIntegration
	recipient uuid.UUID
	message   *integrations.DirectMessage
	sendErr   error
}

func (s *reviewMessageSender) SendDirectMessage(_ context.Context, recipient *ent.User, message *integrations.DirectMessage) error {
	s.recipient = recipient.ID
	s.message = message
	return s.sendErr
}
