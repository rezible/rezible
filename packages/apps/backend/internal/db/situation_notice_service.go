package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/riverqueue/river"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	in "github.com/rezible/rezible/ent/integration"
	sit "github.com/rezible/rezible/ent/situation"
	sitnot "github.com/rezible/rezible/ent/situationnotice"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/messages"
)

const (
	// NoticeSyncDelay is how long after a situation changes its notices are synced. The change is published
	// inside its transaction, so the delay lets that commit before the sync reads it.
	NoticeSyncDelay = 5 * time.Second
	// MaxNoticesPerHour is how many new notices one installation posts to one channel in an hour.
	MaxNoticesPerHour = 10

	noticeCapReason         = "hourly notice limit reached"
	noticeBuildFailedReason = "could not build the notice"
	noticeFailedReason      = "delivery failed"
)

// errNoticeBuild marks a failure to build a notice's content, as opposed to a failure to deliver it.
var errNoticeBuild = errors.New(noticeBuildFailedReason)

// SituationNoticeService posts a notice of each raised situation through every installation that sends them,
// and edits it as the situation changes. Integrations only render and send; which notices exist, when they
// are sent, the hourly limit and failures are decided here.
type SituationNoticeService struct {
	db           rez.Database
	cfg          rez.Config
	clock        rez.Clock
	jobs         rez.JobService
	msgs         rez.MessageQueue
	briefs       rez.SituationBriefService
	outputs      rez.InvestigationOutputService
	integrations rez.IntegrationService
	orgs         rez.OrganizationService
}

func NewSituationNoticeService(db rez.Database, cfg rez.Config, clock rez.Clock, jobService rez.JobService, msgs rez.MessageQueue, briefs rez.SituationBriefService, outputs rez.InvestigationOutputService, integrationService rez.IntegrationService, orgs rez.OrganizationService) *SituationNoticeService {
	return &SituationNoticeService{
		db:           db,
		cfg:          cfg,
		clock:        clock,
		jobs:         jobService,
		msgs:         msgs,
		briefs:       briefs,
		outputs:      outputs,
		integrations: integrationService,
		orgs:         orgs,
	}
}

func (s *SituationNoticeService) MessageHandlers() []rez.MessageEventHandler {
	return []rez.MessageEventHandler{
		messages.NewEventHandler("db.SituationNoticeService.onSituationUpdated", s.onSituationUpdated),
	}
}

func NewSyncSituationNoticesWorker(service *SituationNoticeService) jobs.WorkerDefinition {
	return jobs.DefineWorker[jobs.SyncSituationNotices](&syncSituationNoticesWorker{service: service})
}

type syncSituationNoticesWorker struct {
	river.WorkerDefaults[jobs.SyncSituationNotices]
	service *SituationNoticeService
}

func (w *syncSituationNoticesWorker) Work(ctx context.Context, job *river.Job[jobs.SyncSituationNotices]) error {
	lastAttempt := job.Attempt >= job.MaxAttempts
	return w.service.sync(ctx, job.Args.SituationID, lastAttempt)
}

// onSituationUpdated decides nothing: it can run before the change commits, so it only schedules the sync,
// which reads the situation when it runs.
func (s *SituationNoticeService) onSituationUpdated(ctx context.Context, ev *rez.EventOnSituationUpdated) error {
	dueAt := s.clock.Now().Add(NoticeSyncDelay).Truncate(time.Second).Add(time.Second)
	args := jobs.SyncSituationNotices{SituationID: ev.SituationID, DueAt: dueAt}
	if _, insertErr := s.jobs.Insert(ctx, args, &river.InsertOpts{ScheduledAt: dueAt}); insertErr != nil {
		return fmt.Errorf("insert situation notice sync job: %w", insertErr)
	}
	return nil
}

func (s *SituationNoticeService) RetrySituationNotices(ctx context.Context, situationID uuid.UUID) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		querySituation := tx.Situation.Query().
			Where(sit.ID(situationID))
		exists, existsErr := querySituation.Exist(ctx)
		if existsErr != nil {
			return fmt.Errorf("check situation: %w", existsErr)
		}
		if !exists {
			return fmt.Errorf("%w: situation not found", errs.ErrNotFound)
		}
		retryFailed := tx.SituationNotice.Update().
			Where(sitnot.SituationID(situationID), sitnot.StatusEQ(sitnot.StatusFailed)).
			SetStatus(sitnot.StatusPending).
			ClearError()
		if retryErr := retryFailed.Exec(ctx); retryErr != nil {
			return fmt.Errorf("mark failed situation notices pending: %w", retryErr)
		}
		if publishErr := s.msgs.Publish(ctx, rez.EventOnSituationUpdated{SituationID: situationID}); publishErr != nil {
			return fmt.Errorf("publish situation updated: %w", publishErr)
		}
		return nil
	})
}

// sync creates the notices a raised situation is due, then posts or edits each with the current content.
// A failure is stored on the notices when it is the last attempt.
func (s *SituationNoticeService) sync(ctx context.Context, situationID uuid.UUID, lastAttempt bool) error {
	notices, prepareErr := s.prepare(ctx, situationID)
	if prepareErr != nil {
		return s.failBuild(ctx, situationID, prepareErr, lastAttempt)
	}
	var deliveryErrs []error
	for _, notice := range notices {
		deliverErr := s.deliver(ctx, notice, lastAttempt)
		if deliverErr == nil {
			continue
		}
		if errors.Is(deliverErr, errNoticeBuild) {
			return s.failBuild(ctx, situationID, deliverErr, lastAttempt)
		}
		deliveryErrs = append(deliveryErrs, fmt.Errorf("notice %s: %w", notice.ID, deliverErr))
	}
	return errors.Join(deliveryErrs...)
}

// prepare creates the notices a raised situation is due and returns those to deliver. Only an open situation
// gets new notices: one raised before an installation chose its channel is announced at its next change, a
// closed one never.
func (s *SituationNoticeService) prepare(ctx context.Context, situationID uuid.UUID) ([]*ent.SituationNotice, error) {
	situation, getErr := s.db.Client(ctx).Situation.Get(ctx, situationID)
	if getErr != nil {
		if ent.IsNotFound(getErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("get situation: %w", getErr)
	}
	if situation.RaisedAt == nil {
		return nil, nil
	}
	if situation.ClosedAt == nil {
		if createErr := s.createNotices(ctx, situationID); createErr != nil {
			return nil, createErr
		}
	}

	queryNotices := s.db.Client(ctx).SituationNotice.Query().
		Where(sitnot.SituationID(situationID), sitnot.StatusNEQ(sitnot.StatusSkipped))
	notices, queryErr := queryNotices.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query situation notices: %w", queryErr)
	}
	return notices, nil
}

// failBuild returns the failure to build the content, and on the last attempt stores it on every notice of
// the situation that was not skipped. A posted notice keeps its message and link, so it shows that its
// update failed.
func (s *SituationNoticeService) failBuild(ctx context.Context, situationID uuid.UUID, buildErr error, lastAttempt bool) error {
	if !lastAttempt {
		return buildErr
	}
	failNotices := s.db.Client(ctx).SituationNotice.Update().
		Where(sitnot.SituationID(situationID), sitnot.StatusNEQ(sitnot.StatusSkipped)).
		SetStatus(sitnot.StatusFailed).
		SetError(noticeBuildFailedReason).
		SetLastAttemptAt(s.clock.Now())
	if failErr := failNotices.Exec(ctx); failErr != nil {
		return errors.Join(buildErr, fmt.Errorf("fail situation notices: %w", failErr))
	}
	return buildErr
}

// sender is the installation's notice sender, or nil when it sends none: its integration does not, or is no
// longer available.
func (s *SituationNoticeService) sender(installation *ent.Integration) (integrations.SituationNoticeSender, error) {
	installed, installedErr := s.integrations.AsInstalledIntegration(installation)
	if errors.Is(installedErr, integrations.ErrUnknownIntegration) {
		return nil, nil
	}
	if installedErr != nil {
		return nil, fmt.Errorf("installed integration %s: %w", installation.ID, installedErr)
	}
	sender, supported := installed.(integrations.SituationNoticeSender)
	if !supported {
		return nil, nil
	}
	return sender, nil
}

// createNotices creates a notice for each installation with a notice channel and none for the situation yet.
// Changing the channel later does not move a notice.
func (s *SituationNoticeService) createNotices(ctx context.Context, situationID uuid.UUID) error {
	installations, listErr := s.integrations.ListInstallations(ctx)
	if listErr != nil {
		return fmt.Errorf("list installations: %w", listErr)
	}
	queryNoticed := s.db.Client(ctx).SituationNotice.Query().
		Where(sitnot.SituationID(situationID)).
		Select(sitnot.FieldIntegrationID)
	noticed, noticedErr := queryNoticed.All(ctx)
	if noticedErr != nil {
		return fmt.Errorf("query situation notices: %w", noticedErr)
	}
	noticedIntegrations := mapset.NewSet[uuid.UUID]()
	for _, notice := range noticed {
		noticedIntegrations.Add(notice.IntegrationID)
	}
	for _, installation := range installations {
		if noticedIntegrations.Contains(installation.ID) {
			continue
		}
		sender, senderErr := s.sender(installation)
		if senderErr != nil {
			slog.ErrorContext(ctx, "resolve situation notice sender", "error", senderErr, "integration_id", installation.ID)
			continue
		}
		if sender == nil || sender.SituationNoticeChannel() == "" {
			continue
		}
		if createErr := s.createNotice(ctx, situationID, installation.ID, sender.SituationNoticeChannel()); createErr != nil {
			return createErr
		}
	}
	return nil
}

// createNotice creates the situation's pending notice for the installation. The hourly limit is applied when
// it is first posted.
func (s *SituationNoticeService) createNotice(ctx context.Context, situationID, integrationID uuid.UUID, channelID string) error {
	now := s.clock.Now()
	createNotice := s.db.Client(ctx).SituationNotice.Create().
		SetSituationID(situationID).
		SetIntegrationID(integrationID).
		SetChannelID(channelID).
		SetCreatedAt(now).
		SetUpdatedAt(now)
	upsertNotice := createNotice.
		OnConflictColumns(sitnot.FieldTenantID, sitnot.FieldSituationID, sitnot.FieldIntegrationID).
		DoNothing()
	if createErr := upsertNotice.Exec(ctx); createErr != nil {
		return fmt.Errorf("create situation notice: %w", createErr)
	}
	return nil
}

// capReached is whether the notice's channel has had MaxNoticesPerHour notices first posted by its
// installation in the past hour.
func (s *SituationNoticeService) capReached(ctx context.Context, tx *ent.Client, notice *ent.SituationNotice) (bool, error) {
	// Two first posts to one installation could both count its channel under the limit. Locking the
	// installation row serializes first posts; it is taken after the notice row, and only here.
	lockInstallation := tx.Integration.Query().
		Where(in.ID(notice.IntegrationID)).
		ForUpdate(ent.KeyUpdateLock)
	if _, lockErr := lockInstallation.Only(ctx); lockErr != nil {
		return false, fmt.Errorf("lock installation: %w", lockErr)
	}
	countRecent := tx.SituationNotice.Query().
		Where(
			sitnot.IntegrationID(notice.IntegrationID),
			sitnot.ChannelID(notice.ChannelID),
			sitnot.FirstPostedAtGT(s.clock.Now().Add(-time.Hour)),
		)
	recent, countErr := countRecent.Count(ctx)
	if countErr != nil {
		return false, fmt.Errorf("count recent notices: %w", countErr)
	}
	return recent >= MaxNoticesPerHour, nil
}

// readContent reads what the situation's notice shows, as it is now. The deployment is left out, not the
// notice, when its lookup fails.
func (s *SituationNoticeService) readContent(ctx context.Context, situationID uuid.UUID) (*integrations.SituationNoticeContent, error) {
	querySituation := s.db.Client(ctx).Situation.Query().
		Where(sit.ID(situationID)).
		WithLinks(func(q *ent.SituationLinkQuery) {
			q.WithLinkedSituation()
		}).
		WithInvestigation()
	situation, getErr := querySituation.Only(ctx)
	if getErr != nil {
		return nil, fmt.Errorf("get situation: %w", getErr)
	}
	content := &integrations.SituationNoticeContent{Situation: situation, Now: s.clock.Now()}

	situationURL, urlErr := s.cfg.App.GetFrontendUrl("situations", situation.Reference())
	if urlErr != nil {
		return nil, fmt.Errorf("situation url: %w", urlErr)
	}
	content.SituationURL = situationURL.String()
	if link := content.MergedInto(); link != nil {
		mergedURL, mergedErr := s.cfg.App.GetFrontendUrl("situations", link.Edges.LinkedSituation.Reference())
		if mergedErr != nil {
			return nil, fmt.Errorf("merged situation url: %w", mergedErr)
		}
		content.MergedIntoURL = mergedURL.String()
	}

	if incidentErr := s.readIncidentLinks(ctx, content); incidentErr != nil {
		return nil, incidentErr
	}

	services, servicesErr := s.briefs.ListBriefServices(ctx, situation.ID)
	if servicesErr != nil {
		return nil, fmt.Errorf("list situation services: %w", servicesErr)
	}
	content.Services = services

	if integrations.CheckSources(ctx, s.integrations).Changes.Available {
		deployment, deploymentErr := s.briefs.LatestDeploymentBefore(ctx, situation.ID)
		if deploymentErr != nil {
			slog.ErrorContext(ctx, "look up the latest deployment for a situation notice",
				"error", deploymentErr,
				"situation_id", situation.ID,
			)
		} else {
			content.Deployment = deployment
		}
	}

	if link := situation.Edges.Investigation; link != nil {
		params := rez.GetInvestigationReportParams{
			InvestigationID: link.InvestigationID,
			Selection:       rez.InvestigationReportSelectionCompleted,
		}
		report, reportErr := s.outputs.GetInvestigationReport(ctx, params)
		if reportErr != nil {
			return nil, fmt.Errorf("get investigation report: %w", reportErr)
		}
		content.Report = report
	}
	return content, nil
}

// noticeDelivery is a notice as delivery left it, and why delivery failed, if it did.
type noticeDelivery struct {
	notice *ent.SituationNotice
	err    error
}

// readIncidentLinks offers the newest unresolved incident the situation is part of, or else declaring one, when
// incident management is on and the situation was not merged into another.
func (s *SituationNoticeService) readIncidentLinks(ctx context.Context, content *integrations.SituationNoticeContent) error {
	client := s.db.Client(ctx)
	queryIncident := client.Incident.Query().
		Where(
			incident.HasSituationsWith(sit.ID(content.Situation.ID)),
			incident.ResponseStateNEQ(incident.ResponseStateResolved),
		).
		Order(incident.ByOpenedAt(sql.OrderDesc()), incident.ByID(sql.OrderDesc()))
	unresolved, incidentErr := queryIncident.First(ctx)
	if incidentErr != nil && !ent.IsNotFound(incidentErr) {
		return fmt.Errorf("query unresolved incident: %w", incidentErr)
	}
	if unresolved != nil {
		incidentURL, urlErr := s.cfg.App.GetFrontendUrl("incidents", unresolved.Reference())
		if urlErr != nil {
			return fmt.Errorf("incident url: %w", urlErr)
		}
		content.IncidentURL = incidentURL.String()
		return nil
	}

	situation := content.Situation
	if situation.CloseReason != nil && *situation.CloseReason == sit.CloseReasonMerged {
		return nil
	}
	enabled, enabledErr := s.orgs.IsIncidentManagementEnabled(ctx)
	if enabledErr != nil {
		return fmt.Errorf("read incident management: %w", enabledErr)
	}
	if !enabled {
		return nil
	}
	declareURL, urlErr := s.cfg.App.GetFrontendUrl("situations", situation.Reference())
	if urlErr != nil {
		return fmt.Errorf("declare incident url: %w", urlErr)
	}
	declareURL.RawQuery = url.Values{"declare": {"incident"}}.Encode()
	content.DeclareIncidentURL = declareURL.String()
	return nil
}

// deliver posts the notice, or edits it once posted, then fills in its link and channel name. A notice whose
// installation no longer sends notices is left as it is.
func (s *SituationNoticeService) deliver(ctx context.Context, notice *ent.SituationNotice, lastAttempt bool) error {
	installation, installationErr := s.db.Client(ctx).Integration.Get(ctx, notice.IntegrationID)
	if installationErr != nil {
		return fmt.Errorf("get notice installation: %w", installationErr)
	}
	sender, senderErr := s.sender(installation)
	if senderErr != nil {
		return senderErr
	}
	if sender == nil {
		return nil
	}
	delivery, txErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*noticeDelivery, error) {
		// The post or edit runs while the notice row is locked: two syncs of one situation can overlap, and
		// both would post a notice that has no message yet. The content is read under the lock, so the
		// last sync to hold it sends the situation as it is then.
		lockNotice := tx.SituationNotice.Query().
			Where(sitnot.ID(notice.ID)).
			ForUpdate(ent.KeyUpdateLock)
		locked, lockErr := lockNotice.Only(ctx)
		if lockErr != nil {
			return nil, fmt.Errorf("lock situation notice: %w", lockErr)
		}
		if locked.Status == sitnot.StatusSkipped {
			return &noticeDelivery{notice: locked}, nil
		}
		content, contentErr := s.readContent(ctx, locked.SituationID)
		if contentErr != nil {
			return nil, fmt.Errorf("%w: %w", errNoticeBuild, contentErr)
		}

		update := tx.SituationNotice.UpdateOneID(locked.ID).
			SetLastAttemptAt(s.clock.Now())
		if locked.MessageRef == "" {
			capped, capErr := s.capReached(ctx, tx, locked)
			if capErr != nil {
				return nil, capErr
			}
			if capped {
				update.SetStatus(sitnot.StatusSkipped).SetError(noticeCapReason)
				saved, saveErr := update.Save(ctx)
				if saveErr != nil {
					return nil, fmt.Errorf("skip situation notice: %w", saveErr)
				}
				return &noticeDelivery{notice: saved}, nil
			}
		}

		sendErr := s.send(ctx, sender, locked, content, update)
		if sendErr == nil {
			update.SetStatus(sitnot.StatusPosted).ClearError()
		} else if lastAttempt {
			update.SetStatus(sitnot.StatusFailed).SetError(s.deliveryErrorText(sendErr))
		}
		saved, saveErr := update.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save situation notice: %w", saveErr)
		}
		return &noticeDelivery{notice: saved, err: sendErr}, nil
	})
	if txErr != nil {
		return fmt.Errorf("deliver situation notice: %w", txErr)
	}
	if delivery.err != nil {
		return delivery.err
	}
	if delivery.notice.Status == sitnot.StatusPosted {
		s.fillDetails(ctx, sender, delivery.notice)
	}
	return nil
}

// send posts the notice when it has no message, otherwise edits it.
func (s *SituationNoticeService) send(ctx context.Context, sender integrations.SituationNoticeSender, notice *ent.SituationNotice, content *integrations.SituationNoticeContent, update *ent.SituationNoticeUpdateOne) error {
	if notice.MessageRef != "" {
		return sender.UpdateSituationNotice(ctx, notice.ChannelID, notice.MessageRef, content)
	}
	messageRef, postErr := sender.PostSituationNotice(ctx, notice.ChannelID, content)
	if postErr != nil {
		return postErr
	}
	update.SetMessageRef(messageRef).SetFirstPostedAt(s.clock.Now())
	return nil
}

// fillDetails reads the posted notice's missing link and channel name. They do not fail the delivery: a
// later sync fills in what cannot be read now.
func (s *SituationNoticeService) fillDetails(ctx context.Context, sender integrations.SituationNoticeSender, notice *ent.SituationNotice) {
	if notice.Permalink != "" && notice.ChannelName != "" {
		return
	}
	update := s.db.Client(ctx).SituationNotice.UpdateOneID(notice.ID)
	if notice.Permalink == "" {
		permalink, permalinkErr := sender.SituationNoticeLink(ctx, notice.ChannelID, notice.MessageRef)
		if permalinkErr != nil {
			slog.WarnContext(ctx, "get situation notice link", "error", permalinkErr, "situation_id", notice.SituationID)
		} else {
			update.SetPermalink(permalink)
		}
	}
	if notice.ChannelName == "" {
		channelName, nameErr := sender.ChatChannelName(ctx, notice.ChannelID)
		if nameErr != nil {
			slog.WarnContext(ctx, "get situation notice channel name", "error", nameErr, "situation_id", notice.SituationID)
		} else {
			update.SetChannelName(channelName)
		}
	}
	if updateErr := update.Exec(ctx); updateErr != nil {
		slog.ErrorContext(ctx, "save situation notice details", "error", updateErr, "situation_id", notice.SituationID)
	}
}

// deliveryErrorText is the reason the provider gave, or a generic one: other errors may hold credentials.
func (s *SituationNoticeService) deliveryErrorText(deliveryErr error) string {
	if delivery, isDelivery := errors.AsType[*integrations.NoticeDeliveryError](deliveryErr); isDelivery {
		return delivery.Reason
	}
	return noticeFailedReason
}
