import type {
	InvestigationNextCheck,
	InvestigationProviderRead,
	InvestigationReportAttributes,
	SituationBrief,
	SituationBriefDeployment,
	SituationBriefMergedChange,
	SituationBriefService,
	SituationUnconfirmedMatch,
} from "$lib/api";
import {
	providerReadRangeLabel,
	providerReadView,
	referenceNumbers,
	type ProviderReadView,
} from "./investigation-outputs";

/** The Brief tab's sections after the header, in the order they are shown. */
export type BriefSection =
	| { kind: "interpretation"; nextChecks: NextCheckView[] }
	| { kind: "signals" }
	| { kind: "changes"; state: ChangesState }
	| { kind: "checked"; state: CheckedState }
	| { kind: "unseen"; state: UnseenState };

/** The brief's load state: its sections that come from it show it while it is not loaded. */
export type BriefLoad = { kind: "loading" } | { kind: "error" } | { kind: "loaded"; brief: SituationBrief };

export type NextCheckView = { key: string; check: string; reason: string };

export type UnconfirmedMatchView = {
	key: string;
	/** "May also be `checkout` (unconfirmed)". */
	label: string;
	name: string;
	rationale: string;
};

export type BriefServiceView = {
	id: string;
	name: string;
	unconfirmedMatches: UnconfirmedMatchView[];
};

/** Where a deployment links: its own page, or the evidence that reported it. */
export type DeploymentLink = { kind: "url"; href: string } | { kind: "evidence"; evidenceId: string };

export type MergedChangeView = {
	id: string;
	title: string;
	/** "#42 by alice". */
	detail: string;
	mergedAt: string;
	href: string | undefined;
	linkLabel: string;
};

export type DeploymentView = {
	id: string;
	title: string;
	deployedAt: string;
	environment: string;
	status: string;
	afterOpened: boolean;
	link: DeploymentLink | undefined;
	mergedChanges: MergedChangeView[];
	mergedChangesCut: boolean;
	/** Why merged changes cannot be linked; undefined when they can. */
	unlinkedReason: string | undefined;
};

export type ServiceChangesView = BriefServiceView & {
	deployments: DeploymentView[];
	limits: string[];
};

export type ChangesWindow = { start: string; end: string; label: string };

export type ChangesState =
	| { kind: "loading" }
	| { kind: "error" }
	| { kind: "no_services" }
	| { kind: "not_reported"; reason: string; services: BriefServiceView[] }
	| { kind: "nothing_found"; window: ChangesWindow; services: BriefServiceView[] }
	| { kind: "deployments"; window: ChangesWindow; services: ServiceChangesView[] };

export type CheckedReadView = { id: string; number: number; read: ProviderReadView };

export type CheckedState =
	{ kind: "no_report" } | { kind: "none_cited" } | { kind: "reads"; reads: CheckedReadView[] };

export type UnseenState =
	| { kind: "loading" }
	| { kind: "error" }
	| { kind: "entries"; entries: { key: string; kind: string; text: string }[] };

const MAX_NEXT_CHECKS = 3;
const MIN_VISIBLE_BAR_PERCENT = 4;

const MERGED_CHANGE_LINK_LABELS: Record<SituationBriefMergedChange["link"], string> = {
	merge_commit: "Deployed commit",
	merged_since_previous_deployment: "Merged since the previous deployment",
	merged_in_lookback: "Merged in the 7 days before; no previous deployment",
};

const UNLINKED_REASONS: Record<SituationBriefDeployment["mergedChangesUnlinked"], string | undefined> = {
	"": undefined,
	no_repository: "The deployment names no repository, so merged changes cannot be linked.",
	repository_not_from_code_forge:
		"No code forge has reported this repository, so merged changes cannot be linked.",
};

/** The Brief tab's sections in order, from the brief and the latest report. */
export function briefSections(
	load: BriefLoad,
	report: InvestigationReportAttributes | undefined
): BriefSection[] {
	return [
		{ kind: "interpretation", nextChecks: nextCheckViews(report?.nextChecks ?? []) },
		{ kind: "signals" },
		{ kind: "changes", state: changesState(load) },
		{ kind: "checked", state: checkedState(report) },
		{ kind: "unseen", state: unseenState(load) },
	];
}

export function nextCheckViews(nextChecks: InvestigationNextCheck[]): NextCheckView[] {
	return nextChecks.slice(0, MAX_NEXT_CHECKS).map((nextCheck, index) => ({
		key: `${index}:${nextCheck.check}`,
		check: nextCheck.check,
		reason: nextCheck.reason,
	}));
}

export function unconfirmedMatchViews(matches: SituationUnconfirmedMatch[]): UnconfirmedMatchView[] {
	return matches.map((match) => ({
		key: match.proposalId,
		label: `May also be ${match.name} (unconfirmed)`,
		name: match.name,
		rationale: match.rationale,
	}));
}

function serviceView(service: SituationBriefService): BriefServiceView {
	return {
		id: service.entityId,
		name: service.displayName,
		unconfirmedMatches: unconfirmedMatchViews(service.unconfirmedMatches),
	};
}

export function changesState(load: BriefLoad): ChangesState {
	if (load.kind !== "loaded") {
		return { kind: load.kind };
	}

	const brief = load.brief;
	const services = brief.services.map(serviceView);
	if (services.length === 0) {
		return { kind: "no_services" };
	}

	const changes = brief.changes;
	if (!changes.available) {
		return { kind: "not_reported", reason: changes.unavailableReason, services };
	}

	const window: ChangesWindow = {
		start: changes.start,
		end: changes.end,
		label: providerReadRangeLabel(changes.start, changes.end),
	};
	const changesByService = new Map(
		changes.services.map((serviceChanges) => [serviceChanges.serviceId, serviceChanges])
	);
	const withChanges: ServiceChangesView[] = services.map((service) => {
		const serviceChanges = changesByService.get(service.id);
		return {
			...service,
			deployments: (serviceChanges?.deployments ?? []).map(deploymentView),
			limits: serviceChanges?.limits ?? [],
		};
	});

	const found = withChanges.some((service) => service.deployments.length > 0);
	if (!found) {
		return { kind: "nothing_found", window, services };
	}
	return { kind: "deployments", window, services: withChanges };
}

export function deploymentView(deployment: SituationBriefDeployment): DeploymentView {
	let link: DeploymentLink | undefined;
	if (deployment.url) {
		link = { kind: "url", href: deployment.url };
	} else if (deployment.latestEvidenceId) {
		link = { kind: "evidence", evidenceId: deployment.latestEvidenceId };
	}

	return {
		id: deployment.id,
		title: deploymentTitle(deployment),
		deployedAt: deployment.deployedAt,
		environment: deployment.environment,
		status: deployment.status,
		afterOpened: deployment.afterOpened,
		link,
		mergedChanges: deployment.mergedChanges.map(mergedChangeView),
		mergedChangesCut: deployment.mergedChangesCut,
		unlinkedReason: UNLINKED_REASONS[deployment.mergedChangesUnlinked],
	};
}

function deploymentTitle(deployment: SituationBriefDeployment) {
	if (deployment.version) {
		return deployment.version;
	}
	if (deployment.sha) {
		return deployment.sha.slice(0, 12);
	}
	return "Deployment";
}

function mergedChangeView(change: SituationBriefMergedChange): MergedChangeView {
	let detail = `#${change.number}`;
	if (change.author) {
		detail = `${detail} by ${change.author}`;
	}

	let href: string | undefined;
	if (change.url) {
		href = change.url;
	}

	return {
		id: change.id,
		title: change.title,
		detail,
		mergedAt: change.mergedAt,
		href,
		linkLabel: MERGED_CHANGE_LINK_LABELS[change.link],
	};
}

/** The latest report's provider read citations, in the order it cites them. */
export function checkedState(report: InvestigationReportAttributes | undefined): CheckedState {
	if (!report) {
		return { kind: "no_report" };
	}
	const numbers = referenceNumbers(report.references);
	const reads: CheckedReadView[] = [];
	for (const reference of report.references) {
		const number = numbers.get(reference.id);
		const seen = reads.some((read) => read.id === reference.id);
		if (reference.kind !== "provider_read" || !reference.providerRead || number === undefined || seen) {
			continue;
		}
		reads.push({ id: reference.id, number, read: providerReadView(reference.providerRead) });
	}
	if (reads.length === 0) {
		return { kind: "none_cited" };
	}
	return { kind: "reads", reads };
}

export function unseenState(load: BriefLoad): UnseenState {
	if (load.kind !== "loaded") {
		return { kind: load.kind };
	}
	const entries = load.brief.unseen.map((unseen, index) => ({
		key: `${index}:${unseen.kind}`,
		kind: unseen.kind,
		text: unseen.text,
	}));
	return { kind: "entries", entries };
}

export type ProviderReadLineView = {
	key: string;
	time: string;
	line: string;
	/** The line's own labels as `key=value`. */
	labels: string[];
	/** Shown when the entry stands for repeated lines. */
	repeats: string | undefined;
};

export type ProviderReadPointView = { key: string; time: string; value: number; heightPercent: number };

export type ProviderReadResultView =
	| { kind: "failed" }
	| { kind: "unavailable" }
	| { kind: "log_search"; labels: string[]; lines: ProviderReadLineView[] }
	| { kind: "log_count"; points: ProviderReadPointView[]; range: string; step: string | undefined };

function labelPairs(labels: Record<string, string>): string[] {
	return Object.entries(labels)
		.sort(([first], [second]) => first.localeCompare(second))
		.map(([key, value]) => `${key}=${value}`);
}

function stepLabel(seconds: number): string | undefined {
	if (seconds <= 0) {
		return undefined;
	}
	if (seconds % 3600 === 0) {
		return `${seconds / 3600}h`;
	}
	if (seconds % 60 === 0) {
		return `${seconds / 60}m`;
	}
	return `${seconds}s`;
}

/** What an expanded provider read shows: its lines newest first, or its counts over time. */
export function providerReadResultView(read: InvestigationProviderRead): ProviderReadResultView {
	const attributes = read.attributes;
	if (attributes.summary.status === "failed") {
		return { kind: "failed" };
	}

	const search = attributes.logSearch;
	if (search) {
		const lines = [...search.lines].sort(
			(first, second) => Date.parse(second.time) - Date.parse(first.time)
		);
		return {
			kind: "log_search",
			labels: labelPairs(search.labels),
			lines: lines.map((line, index) => {
				let repeats: string | undefined;
				if (line.count > 1) {
					repeats = `×${line.count}`;
				}
				return {
					key: `${index}:${line.time}`,
					time: line.time,
					line: line.line,
					labels: labelPairs(line.labels),
					repeats,
				};
			}),
		};
	}

	const count = attributes.logCount;
	if (count) {
		const highest = Math.max(0, ...count.points.map((point) => point.value));
		const points = count.points.map((point, index) => {
			// A bar with any lines stays visible beside much larger ones.
			let heightPercent = 0;
			if (point.value > 0) {
				heightPercent = Math.max(MIN_VISIBLE_BAR_PERCENT, Math.round((point.value / highest) * 100));
			}
			return { key: `${index}:${point.time}`, time: point.time, value: point.value, heightPercent };
		});
		return {
			kind: "log_count",
			points,
			range: providerReadRangeLabel(attributes.summary.start, attributes.summary.end),
			step: stepLabel(count.stepSeconds),
		};
	}

	return { kind: "unavailable" };
}
