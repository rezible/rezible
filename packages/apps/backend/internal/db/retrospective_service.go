package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/review"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
)

// ClosedReviewGracePeriod also appears as CLOSED_REVIEW_GRACE_SECONDS in the documents server.
const ClosedReviewGracePeriod = 60 * time.Second

const MaxReviewTextLength = 4000

type RetrospectiveService struct {
	db   rez.Database
	jobs rez.JobService
}

func NewRetrospectiveService(db rez.Database, jobService rez.JobService) (*RetrospectiveService, error) {
	return &RetrospectiveService{db: db, jobs: jobService}, nil
}

func (s *RetrospectiveService) GetRetrospective(ctx context.Context, p predicate.Retrospective) (*ent.Retrospective, error) {
	return s.db.Client(ctx).Retrospective.Query().
		Where(p).
		WithOwner().
		WithClosedBy().
		WithReviews(func(q *ent.ReviewQuery) {
			q.WithReviewer().Order(review.ByCreatedAt(), review.ByID())
		}).
		Only(ctx)
}

func (s *RetrospectiveService) CreateForIncident(ctx context.Context, incidentID uuid.UUID) (*ent.Retrospective, error) {
	return ent.WithTxReturning(ctx, s.db, func(txCtx context.Context, tx *ent.Client) (*ent.Retrospective, error) {
		// Serialize retrospective creation with incident resolution.
		queryIncident := tx.Incident.Query().
			Where(incident.ID(incidentID)).
			ForUpdate()
		inc, incidentErr := queryIncident.Only(txCtx)
		if incidentErr != nil {
			return nil, fmt.Errorf("get incident: %w", incidentErr)
		}
		queryExisting := tx.Retrospective.Query().Where(retrospective.IncidentID(incidentID))
		existing, queryErr := queryExisting.Only(txCtx)
		if queryErr == nil {
			return existing, nil
		}
		if !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("get incident retrospective: %w", queryErr)
		}
		if inc.ResponseState != incident.ResponseStateResolved {
			return nil, fmt.Errorf("%w: incident must be resolved before starting a retrospective", errs.ErrConflict)
		}

		createDoc := tx.Document.Create().
			SetContent([]byte("")).
			SetAccessRestricted(false)
		createdDoc, createDocErr := createDoc.Save(txCtx)
		if createDocErr != nil {
			return nil, fmt.Errorf("create doc: %w", createDocErr)
		}

		analysis, createAnalysisErr := tx.SystemAnalysis.Create().Save(txCtx)
		if createAnalysisErr != nil {
			return nil, fmt.Errorf("create system analysis: %w", createAnalysisErr)
		}

		createRetro := tx.Retrospective.Create().
			SetIncidentID(inc.ID).
			SetDocument(createdDoc).
			SetState(retrospective.StateDraft).
			SetNillableOwnerID(inc.DeclaredByID).
			SetSystemAnalysisID(analysis.ID)
		created, createRetroErr := createRetro.Save(txCtx)
		if createRetroErr != nil {
			return nil, fmt.Errorf("create retrospective: %w", createRetroErr)
		}
		return created, nil
	})
}

func (s *RetrospectiveService) GetReportComposition(ctx context.Context, id uuid.UUID) (*rez.RetrospectiveReportComposition, error) {
	return &rez.RetrospectiveReportComposition{}, nil
}

func (s *RetrospectiveService) SetReport(ctx context.Context, id uuid.UUID, params rez.SetRetrospectiveReportParams) (*rez.RetrospectiveReport, error) {
	return &rez.RetrospectiveReport{}, nil
}

// Every transition locks the retrospective before reading reviews, so decisions and closing serialize.
func (s *RetrospectiveService) transition(ctx context.Context, id uuid.UUID, change func(context.Context, *ent.Client, *ent.Retrospective, uuid.UUID) error) (*ent.Retrospective, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Retrospective, error) {
		actor, actorOK := execution.GetContext(ctx).UserID()
		if !actorOK {
			return nil, errs.ErrForbidden
		}
		lock := tx.Retrospective.Query().
			Where(retrospective.ID(id)).
			ForUpdate()
		if _, lockErr := lock.Only(ctx); lockErr != nil {
			return nil, lockErr
		}
		// Load edges only after acquiring the parent row lock.
		retro, getErr := s.GetRetrospective(ctx, retrospective.ID(id))
		if getErr != nil {
			return nil, getErr
		}
		if retro.State == retrospective.StateClosed {
			return nil, errs.New(errs.CodeConflict, "This retrospective is closed.")
		}
		if changeErr := change(ctx, tx, retro, actor); changeErr != nil {
			return nil, changeErr
		}
		return s.GetRetrospective(ctx, retrospective.ID(id))
	})
}

func (s *RetrospectiveService) requireOwner(retro *ent.Retrospective, actor uuid.UUID) error {
	if retro.OwnerID == nil {
		return errs.New(errs.CodeConflict, "This retrospective has no owner. Take ownership first.")
	}
	if *retro.OwnerID != actor {
		return errs.New(errs.CodeForbidden, "Only the retrospective's owner can do this.")
	}
	return nil
}

func (s *RetrospectiveService) requireMember(ctx context.Context, id uuid.UUID) error {
	exists, queryErr := s.db.Client(ctx).User.Query().
		Where(user.ID(id)).
		Exist(ctx)
	if queryErr != nil {
		return queryErr
	}
	if !exists {
		return errs.New(errs.CodeInvalidInput, "Choose a member of this organization.")
	}
	return nil
}

func (s *RetrospectiveService) requireState(retro *ent.Retrospective, state retrospective.State) error {
	if retro.State == state {
		return nil
	}
	if retro.State == retrospective.StateDraft {
		return errs.New(errs.CodeConflict, "This retrospective is a draft.")
	}
	return errs.New(errs.CodeConflict, "This retrospective is "+strings.ReplaceAll(string(retro.State), "_", " ")+".")
}

// SetOwner hands the retrospective to userID. While it has no owner, a member can only make themselves the owner.
func (s *RetrospectiveService) SetOwner(ctx context.Context, id, userID uuid.UUID) (*ent.Retrospective, error) {
	return s.transition(ctx, id, func(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor uuid.UUID) error {
		if retro.OwnerID == nil {
			if userID != actor {
				return errs.New(errs.CodeForbidden, "Take ownership yourself before handing it to someone else.")
			}
		} else if ownerErr := s.requireOwner(retro, actor); ownerErr != nil {
			return ownerErr
		}
		if memberErr := s.requireMember(ctx, userID); memberErr != nil {
			return memberErr
		}
		for _, r := range retro.Edges.Reviews {
			if r.ReviewerID == userID {
				return errs.New(errs.CodeInvalidInput, "A reviewer can't own the retrospective they review.")
			}
		}
		return tx.Retrospective.UpdateOneID(id).
			SetOwnerID(userID).
			Exec(ctx)
	})
}

func (s *RetrospectiveService) createReview(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor, reviewerID uuid.UUID) error {
	if retro.OwnerID != nil && reviewerID == *retro.OwnerID {
		return errs.New(errs.CodeInvalidInput, "The owner can't review their own retrospective.")
	}
	if memberErr := s.requireMember(ctx, reviewerID); memberErr != nil {
		return memberErr
	}
	create := tx.Review.Create().
		SetRetrospectiveID(retro.ID).
		SetRequesterID(actor).
		SetReviewerID(reviewerID).
		SetState(review.StateWaiting)
	created, createErr := create.Save(ctx)
	if createErr != nil {
		return createErr
	}
	_, insertErr := s.jobs.Insert(ctx, jobs.SendReviewRequestMessage{ReviewID: created.ID}, nil)
	return insertErr
}

func (s *RetrospectiveService) RequestReview(ctx context.Context, id uuid.UUID, reviewerIDs []uuid.UUID) (*ent.Retrospective, error) {
	return s.transition(ctx, id, func(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor uuid.UUID) error {
		if ownerErr := s.requireOwner(retro, actor); ownerErr != nil {
			return ownerErr
		}
		if stateErr := s.requireState(retro, retrospective.StateDraft); stateErr != nil {
			return stateErr
		}
		ids := mapset.NewSet(reviewerIDs...)
		if ids.IsEmpty() {
			return errs.New(errs.CodeInvalidInput, "Choose at least one reviewer.")
		}
		for reviewerID := range ids.Iter() {
			if createErr := s.createReview(ctx, tx, retro, actor, reviewerID); createErr != nil {
				return createErr
			}
		}
		return tx.Retrospective.UpdateOneID(id).
			SetState(retrospective.StateInReview).
			Exec(ctx)
	})
}

func (s *RetrospectiveService) AddReviewer(ctx context.Context, id, userID uuid.UUID) (*ent.Retrospective, error) {
	return s.transition(ctx, id, func(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor uuid.UUID) error {
		if ownerErr := s.requireOwner(retro, actor); ownerErr != nil {
			return ownerErr
		}
		if stateErr := s.requireState(retro, retrospective.StateInReview); stateErr != nil {
			return stateErr
		}
		for _, r := range retro.Edges.Reviews {
			if r.ReviewerID == userID {
				return nil
			}
		}
		return s.createReview(ctx, tx, retro, actor, userID)
	})
}

func (s *RetrospectiveService) RemoveReviewer(ctx context.Context, id, userID uuid.UUID) (*ent.Retrospective, error) {
	return s.transition(ctx, id, func(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor uuid.UUID) error {
		if ownerErr := s.requireOwner(retro, actor); ownerErr != nil {
			return ownerErr
		}
		if stateErr := s.requireState(retro, retrospective.StateInReview); stateErr != nil {
			return stateErr
		}
		remove := tx.Review.Delete().
			Where(review.RetrospectiveID(id), review.ReviewerID(userID))
		count, deleteErr := remove.Exec(ctx)
		if deleteErr != nil {
			return deleteErr
		}
		if count > 0 && len(retro.Edges.Reviews) == count {
			return tx.Retrospective.UpdateOneID(id).
				SetState(retrospective.StateDraft).
				Exec(ctx)
		}
		return nil
	})
}

func (s *RetrospectiveService) SetReviewDecision(ctx context.Context, id uuid.UUID, params rez.SetReviewDecisionParams) (*ent.Retrospective, error) {
	return s.transition(ctx, id, func(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor uuid.UUID) error {
		if params.ReviewerID != actor {
			return errs.New(errs.CodeForbidden, "Reviewers record only their own decision.")
		}
		if stateErr := s.requireState(retro, retrospective.StateInReview); stateErr != nil {
			return stateErr
		}
		feedback := strings.TrimSpace(params.Feedback)
		if params.Decision == review.StateRequestChanges && feedback == "" {
			return errs.New(errs.CodeInvalidInput, "Add feedback when requesting changes.")
		}
		if utf8.RuneCountInString(feedback) > MaxReviewTextLength {
			return errs.New(errs.CodeInvalidInput, "Feedback must be 4000 characters or fewer.")
		}
		for _, r := range retro.Edges.Reviews {
			if r.ReviewerID != params.ReviewerID {
				continue
			}
			update := tx.Review.UpdateOneID(r.ID).SetState(params.Decision)
			if params.Decision == review.StateRequestChanges {
				update.SetFeedback(feedback)
			}
			return update.Exec(ctx)
		}
		return errs.New(errs.CodeForbidden, "You haven't been asked to review this retrospective.")
	})
}

func (s *RetrospectiveService) CloseReview(ctx context.Context, id uuid.UUID) (*ent.Retrospective, error) {
	return s.transition(ctx, id, func(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor uuid.UUID) error {
		if ownerErr := s.requireOwner(retro, actor); ownerErr != nil {
			return ownerErr
		}
		waiting := 0
		for _, r := range retro.Edges.Reviews {
			if r.State != review.StateApproved {
				waiting++
			}
		}
		if waiting > 0 {
			reviewers := "reviewers"
			if waiting == 1 {
				reviewers = "reviewer"
			}
			return errs.New(errs.CodeConflict, fmt.Sprintf("Every reviewer must approve before closing. Waiting on %d %s.", waiting, reviewers))
		}
		return s.closeLocked(ctx, tx, retro, actor, nil)
	})
}

func (s *RetrospectiveService) SkipReview(ctx context.Context, id uuid.UUID, reason string) (*ent.Retrospective, error) {
	return s.transition(ctx, id, func(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor uuid.UUID) error {
		if ownerErr := s.requireOwner(retro, actor); ownerErr != nil {
			return ownerErr
		}
		reason = strings.TrimSpace(reason)
		if reason == "" {
			return errs.New(errs.CodeInvalidInput, "Add a reason for skipping.")
		}
		if utf8.RuneCountInString(reason) > MaxReviewTextLength {
			return errs.New(errs.CodeInvalidInput, "The reason must be 4000 characters or fewer.")
		}
		return s.closeLocked(ctx, tx, retro, actor, &reason)
	})
}

func (s *RetrospectiveService) closeLocked(ctx context.Context, tx *ent.Client, retro *ent.Retrospective, actor uuid.UUID, skipReason *string) error {
	update := tx.Retrospective.UpdateOneID(retro.ID).
		SetState(retrospective.StateClosed).
		SetClosedByID(actor).
		SetNillableSkipReason(skipReason)
	if updateErr := update.Exec(ctx); updateErr != nil {
		return updateErr
	}
	// closed_at uses the database clock; Ent can't set an expression without sql/modifier.
	return ent.ExecTx(ctx, "UPDATE retrospectives SET closed_at = now() WHERE id = $1", retro.ID)
}

// The worker holds its own dependencies: integrations depend on RetrospectiveService, so the service can't
// depend on IntegrationService.
func NewSendReviewRequestMessageWorker(db rez.Database, cfg rez.Config, integrationService rez.IntegrationService) jobs.WorkerDefinition {
	worker := &sendReviewRequestMessageWorker{db: db, cfg: cfg, integrations: integrationService}
	return jobs.DefineWorker[jobs.SendReviewRequestMessage](worker)
}

type sendReviewRequestMessageWorker struct {
	river.WorkerDefaults[jobs.SendReviewRequestMessage]
	db           rez.Database
	cfg          rez.Config
	integrations rez.IntegrationService
}

func (w *sendReviewRequestMessageWorker) Work(ctx context.Context, job *river.Job[jobs.SendReviewRequestMessage]) error {
	sendErr := w.send(ctx, job.Args)
	if sendErr == nil || job.Attempt < job.MaxAttempts {
		return sendErr
	}
	update := w.db.Client(ctx).Review.UpdateOneID(job.Args.ReviewID).
		SetMessageFailedAt(time.Now())
	if updateErr := update.Exec(ctx); updateErr != nil && !ent.IsNotFound(updateErr) {
		return errors.Join(sendErr, updateErr)
	}
	return sendErr // The job middleware logs the final failure once.
}

func (w *sendReviewRequestMessageWorker) send(ctx context.Context, args jobs.SendReviewRequestMessage) error {
	query := w.db.Client(ctx).Review.Query().
		Where(review.ID(args.ReviewID)).
		WithRequester().
		WithReviewer().
		WithRetrospective(func(q *ent.RetrospectiveQuery) { q.WithIncident() })
	r, getErr := query.Only(ctx)
	if ent.IsNotFound(getErr) {
		return nil
	}
	if getErr != nil {
		return getErr
	}
	retro := r.Edges.Retrospective
	if r.State != review.StateWaiting || retro == nil || retro.State != retrospective.StateInReview {
		return nil
	}
	inc := retro.Edges.Incident
	if inc == nil {
		slog.DebugContext(ctx, "incident archived, not sending review request", "review_id", r.ID)
		return nil
	}
	link, linkErr := w.cfg.App.GetFrontendUrl("incidents", inc.Slug, "report")
	if linkErr != nil {
		return linkErr
	}
	message := integrations.DirectMessage{
		// Attribute the request to the person who made it, even after an ownership handover.
		Text:    fmt.Sprintf("%s asked you to review the retrospective for %q.", r.Edges.Requester.Name, inc.Title),
		LinkURL: link.String(), LinkLabel: "Open the review",
	}
	installations, listErr := w.integrations.ListInstallations(ctx)
	if listErr != nil {
		return listErr
	}
	var deliveryErrs []error
	for _, installation := range installations {
		installed, installedErr := w.integrations.AsInstalledIntegration(installation)
		if errors.Is(installedErr, integrations.ErrUnknownIntegration) {
			continue
		}
		if installedErr != nil {
			deliveryErrs = append(deliveryErrs, installedErr)
			continue
		}
		sender, supported := installed.(integrations.DirectMessageSender)
		if !supported {
			continue
		}
		sendErr := sender.SendDirectMessage(ctx, r.Edges.Reviewer, &message)
		if errors.Is(sendErr, integrations.ErrNoRecipient) {
			continue
		}
		if sendErr == nil {
			return nil
		}
		deliveryErrs = append(deliveryErrs, sendErr)
	}
	if len(deliveryErrs) > 0 {
		return errors.Join(deliveryErrs...)
	}
	slog.DebugContext(ctx, "no chat recipient for review request", "review_id", r.ID)
	return nil
}
