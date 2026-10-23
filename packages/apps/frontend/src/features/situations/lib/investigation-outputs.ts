import type {
	AgentProviderRead,
	InvestigationFinding,
	InvestigationHypothesis,
	InvestigationReference,
	InvestigationReportAttributes,
} from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { markdownSummary } from "$components/rich-text-view/markdown-summary";
import { hypothesisStatus, invalidatedStatus, provisionalStatus } from "./status";

export type CitationView = {
	/** Knowledge-evidence ID passed to SourceInspection.openEvidence, or a provider read's ID. */
	id: string;
	/** 1-based, stable across the whole page. */
	number: number;
	/** Set when the citation is a provider read rather than knowledge evidence. */
	providerRead: ProviderReadView | undefined;
};

/** What a cited provider read read, and where to open it. */
export type ProviderReadView = {
	label: string;
	serviceName: string;
	/** The integration that ran the read; empty once it is deleted. */
	integration: string;
	/** The range read, with its date: a read can start the day before the situation opened. */
	range: string;
	query: string;
	limits: string[];
	/** Undefined when the read failed before the provider ran it. */
	sourceUrl: string | undefined;
	/** Names the provider the link opens in. */
	sourceLabel: string;
	failed: boolean;
	/** The items the investigation cited, newest citation first. */
	highlight: ProviderReadItemView[];
};

/** A log line or a counted point, shown the same way. */
export type ProviderReadItemView = {
	key: string;
	time: string;
	/** The line's text, or the point's value. */
	content: string;
	/** The item's labels as `key=value`. */
	labels: string[];
	/** Shown when the item stands for repeats. */
	repeats: string | undefined;
};

export type FindingView = {
	id: string;
	number: number;
	title: string;
	bodyMarkdown: string;
	bodySummary: string | undefined;
	provisional: boolean;
	invalidated: boolean;
	relations: string[];
	citations: CitationView[];
};

export type HypothesisView = {
	id: string;
	number: number;
	title: string;
	status: StatusPresentation;
	justificationMarkdown: string;
	justificationSummary: string | undefined;
	provisional: boolean;
	citations: CitationView[];
};

/** Row shape shared by the findings and hypotheses lists. */
export type OutputRow = {
	id: string;
	number: number;
	title: string;
	markdown: string;
	summary: string | undefined;
	badges: StatusPresentation[];
	relations: string[];
	citations: CitationView[];
	muted: boolean;
};

const OUTPUT_SUMMARY_LIMIT = 240;

function isFinding(finding: InvestigationFinding) {
	return finding.attributes.userInputId === null;
}

/**
 * Assigns citation numbers to knowledge-evidence IDs in order of first appearance:
 * report references, then findings (list order), then hypotheses (list order).
 */
export function buildCitationIndex(
	report: InvestigationReportAttributes | undefined,
	findings: InvestigationFinding[],
	hypotheses: InvestigationHypothesis[]
): Map<string, number> {
	const references = [...(report?.references ?? [])];
	for (const finding of findings.filter(isFinding)) {
		references.push(...finding.attributes.references);
	}
	for (const hypothesis of hypotheses) {
		references.push(...hypothesis.attributes.references);
	}
	return referenceNumbers(references);
}

const PROVIDER_READ_LABELS: Record<string, string> = {
	search_logs: "Log search",
	count_logs: "Log count",
};

const PROVIDER_READ_RANGE_FORMAT = new Intl.DateTimeFormat(undefined, {
	month: "short",
	day: "numeric",
	hour: "2-digit",
	minute: "2-digit",
});

export function providerReadRangeLabel(start: string, end: string): string {
	const startDate = new Date(start);
	const endDate = new Date(end);
	if (!Number.isFinite(startDate.getTime()) || !Number.isFinite(endDate.getTime())) {
		return "Range unavailable";
	}
	return PROVIDER_READ_RANGE_FORMAT.formatRange(startDate, endDate);
}

/** Numbers references 1-based in order of first appearance; a repeated reference keeps its first number. */
export function referenceNumbers(references: { id: string }[]): Map<string, number> {
	const numbers = new Map<string, number>();
	for (const reference of references) {
		if (!numbers.has(reference.id)) {
			numbers.set(reference.id, numbers.size + 1);
		}
	}
	return numbers;
}

function labelPairs(labels: Record<string, string>): string[] {
	return Object.entries(labels)
		.sort(([first], [second]) => first.localeCompare(second))
		.map(([key, value]) => `${key}=${value}`);
}

export function providerReadView(read: AgentProviderRead): ProviderReadView {
	const attributes = read.attributes;
	const integration = attributes.integrationDisplayName || attributes.integrationProvider;
	let sourceUrl: string | undefined;
	if (attributes.sourceUrl) {
		sourceUrl = attributes.sourceUrl;
	}

	let sourceLabel = "Open in provider";
	if (integration) {
		sourceLabel = `Open in ${integration}`;
	}

	return {
		label: PROVIDER_READ_LABELS[attributes.toolName] ?? attributes.toolName,
		serviceName: attributes.serviceName,
		integration,
		range: providerReadRangeLabel(attributes.start, attributes.end),
		query: attributes.query,
		limits: attributes.limits,
		sourceUrl,
		sourceLabel,
		failed: attributes.status === "failed",
		highlight: attributes.highlight.map((item) => {
			let repeats: string | undefined;
			if (item.count > 1) {
				repeats = `×${item.count}`;
			}
			let content = item.text ?? "";
			if (item.value !== undefined) {
				content = String(item.value);
			}
			return {
				key: item.ref,
				time: item.time,
				content,
				labels: labelPairs(item.labels),
				repeats,
			};
		}),
	};
}

export function citationViews(
	references: InvestigationReference[],
	citations: Map<string, number>
): CitationView[] {
	const views = new Map<string, CitationView>();
	for (const reference of references) {
		const number = citations.get(reference.id);
		if (number === undefined || views.has(reference.id)) {
			continue;
		}
		let providerRead: ProviderReadView | undefined;
		if (reference.kind === "provider_read" && reference.providerRead) {
			providerRead = providerReadView(reference.providerRead);
		}
		views.set(reference.id, { id: reference.id, number, providerRead });
	}
	return [...views.values()].sort((first, second) => first.number - second.number);
}

function relationLabel(relation: string, number: number) {
	switch (relation) {
		case "supports":
			return `Supports finding ${number}`;
		case "contradicts":
			return `Contradicts finding ${number}`;
		case "invalidates":
			return `Invalidates finding ${number}`;
		default:
			return `Related to finding ${number}`;
	}
}

export function buildFindingViews(
	findings: InvestigationFinding[],
	citations: Map<string, number>
): FindingView[] {
	const visible = findings.filter(isFinding);
	const numbers = new Map<string, number>();
	visible.forEach((finding, index) => numbers.set(finding.id, index + 1));

	return visible.map((finding, index) => {
		const attributes = finding.attributes;
		const relations: string[] = [];
		for (const reference of attributes.findingReferences) {
			const number = numbers.get(reference.versionId);
			if (number !== undefined) {
				relations.push(relationLabel(reference.relation, number));
			}
		}

		return {
			id: finding.id,
			number: index + 1,
			title: attributes.title,
			bodyMarkdown: attributes.body,
			bodySummary: markdownSummary(attributes.body, OUTPUT_SUMMARY_LIMIT),
			provisional: attributes.provisional,
			invalidated: attributes.invalidatedByVersionIds.length > 0,
			relations,
			citations: citationViews(attributes.references, citations),
		};
	});
}

export function buildHypothesisViews(
	hypotheses: InvestigationHypothesis[],
	citations: Map<string, number>
): HypothesisView[] {
	return hypotheses.map((hypothesis, index) => {
		const attributes = hypothesis.attributes;
		return {
			id: hypothesis.id,
			number: index + 1,
			title: attributes.title,
			status: hypothesisStatus(attributes.status),
			justificationMarkdown: attributes.justification,
			justificationSummary: markdownSummary(attributes.justification, OUTPUT_SUMMARY_LIMIT),
			provisional: attributes.provisional,
			citations: citationViews(attributes.references, citations),
		};
	});
}

export function findingRow(finding: FindingView): OutputRow {
	const badges: StatusPresentation[] = [];
	if (finding.invalidated) {
		badges.push(invalidatedStatus);
	}
	if (finding.provisional) {
		badges.push(provisionalStatus);
	}
	return {
		id: finding.id,
		number: finding.number,
		title: finding.title,
		markdown: finding.bodyMarkdown,
		summary: finding.bodySummary,
		badges,
		relations: finding.relations,
		citations: finding.citations,
		muted: finding.invalidated,
	};
}

export function hypothesisRow(hypothesis: HypothesisView): OutputRow {
	const badges: StatusPresentation[] = [hypothesis.status];
	if (hypothesis.provisional) {
		badges.push(provisionalStatus);
	}
	return {
		id: hypothesis.id,
		number: hypothesis.number,
		title: hypothesis.title,
		markdown: hypothesis.justificationMarkdown,
		summary: hypothesis.justificationSummary,
		badges,
		relations: [],
		citations: hypothesis.citations,
		muted: false,
	};
}
