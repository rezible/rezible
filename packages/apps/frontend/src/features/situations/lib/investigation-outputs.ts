import type { InvestigationFinding, InvestigationHypothesis, InvestigationReportAttributes } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { markdownSummary } from "$components/rich-text-view/markdown-summary";
import { hypothesisStatus, invalidatedStatus, provisionalStatus } from "./status";

export type CitationView = {
	/** Knowledge-evidence ID passed to SourceInspection.openEvidence. */
	id: string;
	/** 1-based, stable across the whole page. */
	number: number;
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
	const index = new Map<string, number>();
	const add = (references: { id: string }[]) => {
		for (const reference of references) {
			if (!index.has(reference.id)) {
				index.set(reference.id, index.size + 1);
			}
		}
	};

	add(report?.references ?? []);
	for (const finding of findings.filter(isFinding)) {
		add(finding.attributes.references);
	}
	for (const hypothesis of hypotheses) {
		add(hypothesis.attributes.references);
	}
	return index;
}

export function citationViews(references: { id: string }[], citations: Map<string, number>): CitationView[] {
	const views = new Map<string, CitationView>();
	for (const reference of references) {
		const number = citations.get(reference.id);
		if (number !== undefined && !views.has(reference.id)) {
			views.set(reference.id, { id: reference.id, number });
		}
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
