// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"sync"

	"entgo.io/ent/dialect"
)

// Tx is a transactional client that is created by calling Client.Tx().
type Tx struct {
	config
	// Environment is the client for interacting with the Environment builders.
	Environment *EnvironmentClient
	// Functionality is the client for interacting with the Functionality builders.
	Functionality *FunctionalityClient
	// Incident is the client for interacting with the Incident builders.
	Incident *IncidentClient
	// IncidentDebrief is the client for interacting with the IncidentDebrief builders.
	IncidentDebrief *IncidentDebriefClient
	// IncidentDebriefMessage is the client for interacting with the IncidentDebriefMessage builders.
	IncidentDebriefMessage *IncidentDebriefMessageClient
	// IncidentDebriefQuestion is the client for interacting with the IncidentDebriefQuestion builders.
	IncidentDebriefQuestion *IncidentDebriefQuestionClient
	// IncidentDebriefSuggestion is the client for interacting with the IncidentDebriefSuggestion builders.
	IncidentDebriefSuggestion *IncidentDebriefSuggestionClient
	// IncidentEvent is the client for interacting with the IncidentEvent builders.
	IncidentEvent *IncidentEventClient
	// IncidentEventContext is the client for interacting with the IncidentEventContext builders.
	IncidentEventContext *IncidentEventContextClient
	// IncidentEventContributingFactor is the client for interacting with the IncidentEventContributingFactor builders.
	IncidentEventContributingFactor *IncidentEventContributingFactorClient
	// IncidentEventEvidence is the client for interacting with the IncidentEventEvidence builders.
	IncidentEventEvidence *IncidentEventEvidenceClient
	// IncidentEventSystemComponent is the client for interacting with the IncidentEventSystemComponent builders.
	IncidentEventSystemComponent *IncidentEventSystemComponentClient
	// IncidentField is the client for interacting with the IncidentField builders.
	IncidentField *IncidentFieldClient
	// IncidentFieldOption is the client for interacting with the IncidentFieldOption builders.
	IncidentFieldOption *IncidentFieldOptionClient
	// IncidentLink is the client for interacting with the IncidentLink builders.
	IncidentLink *IncidentLinkClient
	// IncidentMilestone is the client for interacting with the IncidentMilestone builders.
	IncidentMilestone *IncidentMilestoneClient
	// IncidentRole is the client for interacting with the IncidentRole builders.
	IncidentRole *IncidentRoleClient
	// IncidentRoleAssignment is the client for interacting with the IncidentRoleAssignment builders.
	IncidentRoleAssignment *IncidentRoleAssignmentClient
	// IncidentSeverity is the client for interacting with the IncidentSeverity builders.
	IncidentSeverity *IncidentSeverityClient
	// IncidentTag is the client for interacting with the IncidentTag builders.
	IncidentTag *IncidentTagClient
	// IncidentTeamAssignment is the client for interacting with the IncidentTeamAssignment builders.
	IncidentTeamAssignment *IncidentTeamAssignmentClient
	// IncidentType is the client for interacting with the IncidentType builders.
	IncidentType *IncidentTypeClient
	// MeetingSchedule is the client for interacting with the MeetingSchedule builders.
	MeetingSchedule *MeetingScheduleClient
	// MeetingSession is the client for interacting with the MeetingSession builders.
	MeetingSession *MeetingSessionClient
	// OncallAlert is the client for interacting with the OncallAlert builders.
	OncallAlert *OncallAlertClient
	// OncallAlertInstance is the client for interacting with the OncallAlertInstance builders.
	OncallAlertInstance *OncallAlertInstanceClient
	// OncallHandoverTemplate is the client for interacting with the OncallHandoverTemplate builders.
	OncallHandoverTemplate *OncallHandoverTemplateClient
	// OncallRoster is the client for interacting with the OncallRoster builders.
	OncallRoster *OncallRosterClient
	// OncallSchedule is the client for interacting with the OncallSchedule builders.
	OncallSchedule *OncallScheduleClient
	// OncallScheduleParticipant is the client for interacting with the OncallScheduleParticipant builders.
	OncallScheduleParticipant *OncallScheduleParticipantClient
	// OncallUserShift is the client for interacting with the OncallUserShift builders.
	OncallUserShift *OncallUserShiftClient
	// OncallUserShiftAnnotation is the client for interacting with the OncallUserShiftAnnotation builders.
	OncallUserShiftAnnotation *OncallUserShiftAnnotationClient
	// OncallUserShiftCover is the client for interacting with the OncallUserShiftCover builders.
	OncallUserShiftCover *OncallUserShiftCoverClient
	// OncallUserShiftHandover is the client for interacting with the OncallUserShiftHandover builders.
	OncallUserShiftHandover *OncallUserShiftHandoverClient
	// ProviderConfig is the client for interacting with the ProviderConfig builders.
	ProviderConfig *ProviderConfigClient
	// ProviderSyncHistory is the client for interacting with the ProviderSyncHistory builders.
	ProviderSyncHistory *ProviderSyncHistoryClient
	// Retrospective is the client for interacting with the Retrospective builders.
	Retrospective *RetrospectiveClient
	// RetrospectiveDiscussion is the client for interacting with the RetrospectiveDiscussion builders.
	RetrospectiveDiscussion *RetrospectiveDiscussionClient
	// RetrospectiveDiscussionReply is the client for interacting with the RetrospectiveDiscussionReply builders.
	RetrospectiveDiscussionReply *RetrospectiveDiscussionReplyClient
	// RetrospectiveReview is the client for interacting with the RetrospectiveReview builders.
	RetrospectiveReview *RetrospectiveReviewClient
	// SystemAnalysis is the client for interacting with the SystemAnalysis builders.
	SystemAnalysis *SystemAnalysisClient
	// SystemAnalysisComponent is the client for interacting with the SystemAnalysisComponent builders.
	SystemAnalysisComponent *SystemAnalysisComponentClient
	// SystemAnalysisRelationship is the client for interacting with the SystemAnalysisRelationship builders.
	SystemAnalysisRelationship *SystemAnalysisRelationshipClient
	// SystemComponent is the client for interacting with the SystemComponent builders.
	SystemComponent *SystemComponentClient
	// SystemComponentConstraint is the client for interacting with the SystemComponentConstraint builders.
	SystemComponentConstraint *SystemComponentConstraintClient
	// SystemComponentControl is the client for interacting with the SystemComponentControl builders.
	SystemComponentControl *SystemComponentControlClient
	// SystemComponentKind is the client for interacting with the SystemComponentKind builders.
	SystemComponentKind *SystemComponentKindClient
	// SystemComponentSignal is the client for interacting with the SystemComponentSignal builders.
	SystemComponentSignal *SystemComponentSignalClient
	// SystemRelationshipControlAction is the client for interacting with the SystemRelationshipControlAction builders.
	SystemRelationshipControlAction *SystemRelationshipControlActionClient
	// SystemRelationshipFeedbackSignal is the client for interacting with the SystemRelationshipFeedbackSignal builders.
	SystemRelationshipFeedbackSignal *SystemRelationshipFeedbackSignalClient
	// Task is the client for interacting with the Task builders.
	Task *TaskClient
	// Team is the client for interacting with the Team builders.
	Team *TeamClient
	// User is the client for interacting with the User builders.
	User *UserClient

	// lazily loaded.
	client     *Client
	clientOnce sync.Once
	// ctx lives for the life of the transaction. It is
	// the same context used by the underlying connection.
	ctx context.Context
}

type (
	// Committer is the interface that wraps the Commit method.
	Committer interface {
		Commit(context.Context, *Tx) error
	}

	// The CommitFunc type is an adapter to allow the use of ordinary
	// function as a Committer. If f is a function with the appropriate
	// signature, CommitFunc(f) is a Committer that calls f.
	CommitFunc func(context.Context, *Tx) error

	// CommitHook defines the "commit middleware". A function that gets a Committer
	// and returns a Committer. For example:
	//
	//	hook := func(next ent.Committer) ent.Committer {
	//		return ent.CommitFunc(func(ctx context.Context, tx *ent.Tx) error {
	//			// Do some stuff before.
	//			if err := next.Commit(ctx, tx); err != nil {
	//				return err
	//			}
	//			// Do some stuff after.
	//			return nil
	//		})
	//	}
	//
	CommitHook func(Committer) Committer
)

// Commit calls f(ctx, m).
func (f CommitFunc) Commit(ctx context.Context, tx *Tx) error {
	return f(ctx, tx)
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	txDriver := tx.config.driver.(*txDriver)
	var fn Committer = CommitFunc(func(context.Context, *Tx) error {
		return txDriver.tx.Commit()
	})
	txDriver.mu.Lock()
	hooks := append([]CommitHook(nil), txDriver.onCommit...)
	txDriver.mu.Unlock()
	for i := len(hooks) - 1; i >= 0; i-- {
		fn = hooks[i](fn)
	}
	return fn.Commit(tx.ctx, tx)
}

// OnCommit adds a hook to call on commit.
func (tx *Tx) OnCommit(f CommitHook) {
	txDriver := tx.config.driver.(*txDriver)
	txDriver.mu.Lock()
	txDriver.onCommit = append(txDriver.onCommit, f)
	txDriver.mu.Unlock()
}

type (
	// Rollbacker is the interface that wraps the Rollback method.
	Rollbacker interface {
		Rollback(context.Context, *Tx) error
	}

	// The RollbackFunc type is an adapter to allow the use of ordinary
	// function as a Rollbacker. If f is a function with the appropriate
	// signature, RollbackFunc(f) is a Rollbacker that calls f.
	RollbackFunc func(context.Context, *Tx) error

	// RollbackHook defines the "rollback middleware". A function that gets a Rollbacker
	// and returns a Rollbacker. For example:
	//
	//	hook := func(next ent.Rollbacker) ent.Rollbacker {
	//		return ent.RollbackFunc(func(ctx context.Context, tx *ent.Tx) error {
	//			// Do some stuff before.
	//			if err := next.Rollback(ctx, tx); err != nil {
	//				return err
	//			}
	//			// Do some stuff after.
	//			return nil
	//		})
	//	}
	//
	RollbackHook func(Rollbacker) Rollbacker
)

// Rollback calls f(ctx, m).
func (f RollbackFunc) Rollback(ctx context.Context, tx *Tx) error {
	return f(ctx, tx)
}

// Rollback rollbacks the transaction.
func (tx *Tx) Rollback() error {
	txDriver := tx.config.driver.(*txDriver)
	var fn Rollbacker = RollbackFunc(func(context.Context, *Tx) error {
		return txDriver.tx.Rollback()
	})
	txDriver.mu.Lock()
	hooks := append([]RollbackHook(nil), txDriver.onRollback...)
	txDriver.mu.Unlock()
	for i := len(hooks) - 1; i >= 0; i-- {
		fn = hooks[i](fn)
	}
	return fn.Rollback(tx.ctx, tx)
}

// OnRollback adds a hook to call on rollback.
func (tx *Tx) OnRollback(f RollbackHook) {
	txDriver := tx.config.driver.(*txDriver)
	txDriver.mu.Lock()
	txDriver.onRollback = append(txDriver.onRollback, f)
	txDriver.mu.Unlock()
}

// Client returns a Client that binds to current transaction.
func (tx *Tx) Client() *Client {
	tx.clientOnce.Do(func() {
		tx.client = &Client{config: tx.config}
		tx.client.init()
	})
	return tx.client
}

func (tx *Tx) init() {
	tx.Environment = NewEnvironmentClient(tx.config)
	tx.Functionality = NewFunctionalityClient(tx.config)
	tx.Incident = NewIncidentClient(tx.config)
	tx.IncidentDebrief = NewIncidentDebriefClient(tx.config)
	tx.IncidentDebriefMessage = NewIncidentDebriefMessageClient(tx.config)
	tx.IncidentDebriefQuestion = NewIncidentDebriefQuestionClient(tx.config)
	tx.IncidentDebriefSuggestion = NewIncidentDebriefSuggestionClient(tx.config)
	tx.IncidentEvent = NewIncidentEventClient(tx.config)
	tx.IncidentEventContext = NewIncidentEventContextClient(tx.config)
	tx.IncidentEventContributingFactor = NewIncidentEventContributingFactorClient(tx.config)
	tx.IncidentEventEvidence = NewIncidentEventEvidenceClient(tx.config)
	tx.IncidentEventSystemComponent = NewIncidentEventSystemComponentClient(tx.config)
	tx.IncidentField = NewIncidentFieldClient(tx.config)
	tx.IncidentFieldOption = NewIncidentFieldOptionClient(tx.config)
	tx.IncidentLink = NewIncidentLinkClient(tx.config)
	tx.IncidentMilestone = NewIncidentMilestoneClient(tx.config)
	tx.IncidentRole = NewIncidentRoleClient(tx.config)
	tx.IncidentRoleAssignment = NewIncidentRoleAssignmentClient(tx.config)
	tx.IncidentSeverity = NewIncidentSeverityClient(tx.config)
	tx.IncidentTag = NewIncidentTagClient(tx.config)
	tx.IncidentTeamAssignment = NewIncidentTeamAssignmentClient(tx.config)
	tx.IncidentType = NewIncidentTypeClient(tx.config)
	tx.MeetingSchedule = NewMeetingScheduleClient(tx.config)
	tx.MeetingSession = NewMeetingSessionClient(tx.config)
	tx.OncallAlert = NewOncallAlertClient(tx.config)
	tx.OncallAlertInstance = NewOncallAlertInstanceClient(tx.config)
	tx.OncallHandoverTemplate = NewOncallHandoverTemplateClient(tx.config)
	tx.OncallRoster = NewOncallRosterClient(tx.config)
	tx.OncallSchedule = NewOncallScheduleClient(tx.config)
	tx.OncallScheduleParticipant = NewOncallScheduleParticipantClient(tx.config)
	tx.OncallUserShift = NewOncallUserShiftClient(tx.config)
	tx.OncallUserShiftAnnotation = NewOncallUserShiftAnnotationClient(tx.config)
	tx.OncallUserShiftCover = NewOncallUserShiftCoverClient(tx.config)
	tx.OncallUserShiftHandover = NewOncallUserShiftHandoverClient(tx.config)
	tx.ProviderConfig = NewProviderConfigClient(tx.config)
	tx.ProviderSyncHistory = NewProviderSyncHistoryClient(tx.config)
	tx.Retrospective = NewRetrospectiveClient(tx.config)
	tx.RetrospectiveDiscussion = NewRetrospectiveDiscussionClient(tx.config)
	tx.RetrospectiveDiscussionReply = NewRetrospectiveDiscussionReplyClient(tx.config)
	tx.RetrospectiveReview = NewRetrospectiveReviewClient(tx.config)
	tx.SystemAnalysis = NewSystemAnalysisClient(tx.config)
	tx.SystemAnalysisComponent = NewSystemAnalysisComponentClient(tx.config)
	tx.SystemAnalysisRelationship = NewSystemAnalysisRelationshipClient(tx.config)
	tx.SystemComponent = NewSystemComponentClient(tx.config)
	tx.SystemComponentConstraint = NewSystemComponentConstraintClient(tx.config)
	tx.SystemComponentControl = NewSystemComponentControlClient(tx.config)
	tx.SystemComponentKind = NewSystemComponentKindClient(tx.config)
	tx.SystemComponentSignal = NewSystemComponentSignalClient(tx.config)
	tx.SystemRelationshipControlAction = NewSystemRelationshipControlActionClient(tx.config)
	tx.SystemRelationshipFeedbackSignal = NewSystemRelationshipFeedbackSignalClient(tx.config)
	tx.Task = NewTaskClient(tx.config)
	tx.Team = NewTeamClient(tx.config)
	tx.User = NewUserClient(tx.config)
}

// txDriver wraps the given dialect.Tx with a nop dialect.Driver implementation.
// The idea is to support transactions without adding any extra code to the builders.
// When a builder calls to driver.Tx(), it gets the same dialect.Tx instance.
// Commit and Rollback are nop for the internal builders and the user must call one
// of them in order to commit or rollback the transaction.
//
// If a closed transaction is embedded in one of the generated entities, and the entity
// applies a query, for example: Environment.QueryXXX(), the query will be executed
// through the driver which created this transaction.
//
// Note that txDriver is not goroutine safe.
type txDriver struct {
	// the driver we started the transaction from.
	drv dialect.Driver
	// tx is the underlying transaction.
	tx dialect.Tx
	// completion hooks.
	mu         sync.Mutex
	onCommit   []CommitHook
	onRollback []RollbackHook
}

// newTx creates a new transactional driver.
func newTx(ctx context.Context, drv dialect.Driver) (*txDriver, error) {
	tx, err := drv.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return &txDriver{tx: tx, drv: drv}, nil
}

// Tx returns the transaction wrapper (txDriver) to avoid Commit or Rollback calls
// from the internal builders. Should be called only by the internal builders.
func (tx *txDriver) Tx(context.Context) (dialect.Tx, error) { return tx, nil }

// Dialect returns the dialect of the driver we started the transaction from.
func (tx *txDriver) Dialect() string { return tx.drv.Dialect() }

// Close is a nop close.
func (*txDriver) Close() error { return nil }

// Commit is a nop commit for the internal builders.
// User must call `Tx.Commit` in order to commit the transaction.
func (*txDriver) Commit() error { return nil }

// Rollback is a nop rollback for the internal builders.
// User must call `Tx.Rollback` in order to rollback the transaction.
func (*txDriver) Rollback() error { return nil }

// Exec calls tx.Exec.
func (tx *txDriver) Exec(ctx context.Context, query string, args, v any) error {
	return tx.tx.Exec(ctx, query, args, v)
}

// Query calls tx.Query.
func (tx *txDriver) Query(ctx context.Context, query string, args, v any) error {
	return tx.tx.Query(ctx, query, args, v)
}

var _ dialect.Driver = (*txDriver)(nil)
