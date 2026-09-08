import type { Situation, SituationAlertEpisode } from "$lib/api";

export type SituationSignalState = "active" | "quiet" | "none";

export type SituationAttentionReason =
	| "report-stale"
	| "no-investigation"
	| "signals-quiet-unassessed"
	| "investigation-running";

export type SituationPresentation = {
	/** Signal activity, derived only from alert episode state — never from situation assessment. */
	signalState: SituationSignalState;
	openSignalCount: number;
	totalSignalCount: number;
	lastObservedAt: string | undefined;
	latestSignal: SituationAlertEpisode | undefined;
	hasInvestigation: boolean;
	/** An investigation turn has been requested but not yet produced a report. */
	investigationRunning: boolean;
	/** The completed report covers the current evidence revision. */
	reportFresh: boolean;
	/** The completed report covers an older evidence revision than the current one. */
	reportStale: boolean;
	hasReport: boolean;
	reportLimitations: string[];
	attentionReasons: SituationAttentionReason[];
	/** Higher sorts earlier in the Needs Attention view. */
	attentionPriority: number;
};

const openEpisodes = (situation: Situation) =>
	situation.attributes.alertEpisodes.filter((episode) => episode.attributes.status === "open");

const latestObserved = (situation: Situation): SituationAlertEpisode | undefined => {
	return [...situation.attributes.alertEpisodes].sort(
		(a, b) =>
			new Date(b.attributes.lastObservedAt).getTime() -
			new Date(a.attributes.lastObservedAt).getTime()
	)[0];
};

/**
 * Derives dashboard presentation state from explicit situation fields.
 * Signal activity is episode state, not an assessment: a situation whose
 * signals have gone quiet is never treated as resolved or safe here.
 */
export const deriveSituationPresentation = (situation: Situation): SituationPresentation => {
	const { attributes: attrs } = situation;
	const episodes = attrs.alertEpisodes;
	const open = openEpisodes(situation);
	const signalState: SituationSignalState =
		open.length > 0 ? "active" : episodes.length > 0 ? "quiet" : "none";

	const investigation = attrs.investigation;
	const report = investigation?.attributes.report;
	const hasReport = !!report;
	const investigationRunning =
		!!investigation &&
		investigation.attributes.requestedRevision > investigation.attributes.completedRevision;
	const reportStale =
		!!investigation &&
		!investigationRunning &&
		investigation.attributes.completedRevision < investigation.attributes.evidenceRevision;

	const attentionReasons: SituationAttentionReason[] = [];
	if (reportStale) attentionReasons.push("report-stale");
	if (attrs.status === "open") {
		if (!investigation) attentionReasons.push("no-investigation");
		if (signalState === "quiet" && !hasReport) {
			attentionReasons.push("signals-quiet-unassessed");
		}
	}
	if (investigationRunning) attentionReasons.push("investigation-running");

	const priorityOf: Record<SituationAttentionReason, number> = {
		"report-stale": 3,
		"no-investigation": 2,
		"signals-quiet-unassessed": 1,
		"investigation-running": 0,
	};
	const attentionPriority = Math.max(0, ...attentionReasons.map((reason) => priorityOf[reason]));

	return {
		signalState,
		openSignalCount: open.length,
		totalSignalCount: episodes.length,
		lastObservedAt: latestObserved(situation)?.attributes.lastObservedAt,
		latestSignal: latestObserved(situation),
		hasInvestigation: !!investigation,
		investigationRunning,
		reportFresh: !!investigation && !reportStale && hasReport,
		reportStale,
		hasReport,
		reportLimitations: report?.limitations ?? [],
		attentionReasons,
		attentionPriority,
	};
};

export const attentionReasonLabels: Record<SituationAttentionReason, string> = {
	"report-stale": "Report is out of date with current evidence",
	"no-investigation": "No investigation yet",
	"signals-quiet-unassessed": "Safety not assessed",
	"investigation-running": "Investigation in progress",
};

export const byAttentionPriority = (a: Situation, b: Situation): number => {
	const diff =
		deriveSituationPresentation(b).attentionPriority -
		deriveSituationPresentation(a).attentionPriority;
	if (diff !== 0) return diff;
	return new Date(b.attributes.openedAt).getTime() - new Date(a.attributes.openedAt).getTime();
};
