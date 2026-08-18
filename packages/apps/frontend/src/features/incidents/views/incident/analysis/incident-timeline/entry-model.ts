import type {
	CreateSystemAnalysisEntryAttributes,
	SystemAnalysisEntry,
	SystemAnalysisEntryAttributes,
	SystemAnalysisEntrySubject,
} from "$lib/api";

type EntryProperties = Record<string, unknown>;

export type TimelineEntryDecisionContext = {
	optionsConsidered: string[];
	constraints: string[];
	decisionRationale: string;
};

export type TimelineEntryContributingFactor = {
	id: string;
	attributes: {
		factorTypeId: string;
		description: string;
		links: string[];
	};
};

export type TimelineEntryEvidenceAttributes = {
	source: string;
	value: string;
};

export type TimelineEntryEvidence = {
	id: string;
	attributes: TimelineEntryEvidenceAttributes;
};

export type TimelineEntrySystemContextAttributes = {
	systemAnalysisNodeId?: string;
	knowledgeEntityId: string;
	relationship: string;
};

export type TimelineEntrySystemContext = {
	id?: string;
	attributes: TimelineEntrySystemContextAttributes;
};

export type TimelineAnalysisEntryAttributes = {
	kind: SystemAnalysisEntryAttributes["kind"];
	title: string;
	description?: string;
	timestamp: string;
	isKey: boolean;
	decisionContext: TimelineEntryDecisionContext;
	contributingFactors: TimelineEntryContributingFactor[];
	evidence: TimelineEntryEvidence[];
	systemContext: TimelineEntrySystemContext[];
};

export type TimelineAnalysisEntry = {
	id: string;
	attributes: TimelineAnalysisEntryAttributes;
};

export const makeDefaultDecisionContext = (): TimelineEntryDecisionContext => ({
	optionsConsidered: [],
	constraints: [],
	decisionRationale: "",
});

const isRecord = (value: unknown): value is EntryProperties =>
	typeof value === "object" && value !== null && !Array.isArray(value);

const isStringArray = (value: unknown): value is string[] =>
	Array.isArray(value) && value.every((item) => typeof item === "string");

const getDecisionContext = (properties: EntryProperties): TimelineEntryDecisionContext => {
	const value = properties.decisionContext;
	if (!isRecord(value)) return makeDefaultDecisionContext();

	return {
		optionsConsidered: isStringArray(value.optionsConsidered) ? value.optionsConsidered : [],
		constraints: isStringArray(value.constraints) ? value.constraints : [],
		decisionRationale: typeof value.decisionRationale === "string" ? value.decisionRationale : "",
	};
};

const getArrayProperty = <T>(properties: EntryProperties, key: string): T[] => {
	const value = properties[key];
	return Array.isArray(value) ? (value as T[]) : [];
};

const systemContextFromSubject = (
	subject: SystemAnalysisEntrySubject
): TimelineEntrySystemContext | undefined => {
	const attrs = subject.attributes;
	if (attrs.subjectKind !== "entity") return undefined;

	return {
		id: subject.id,
		attributes: {
			knowledgeEntityId: attrs.subjectId,
			relationship: attrs.role,
		},
	};
};

export const systemAnalysisEntryToTimelineEntry = (
	entry: SystemAnalysisEntry
): TimelineAnalysisEntry | undefined => {
	const attrs = entry.attributes;
	if (!attrs.occurredAt) return undefined;

	const properties = isRecord(attrs.properties) ? attrs.properties : {};
	const systemContext = attrs.subjects
		.map(systemContextFromSubject)
		.filter((context): context is TimelineEntrySystemContext => !!context);

	return {
		id: entry.id,
		attributes: {
			kind: attrs.kind,
			title: attrs.title,
			description: attrs.body,
			timestamp: attrs.occurredAt,
			isKey: properties.isKey === true,
			decisionContext: getDecisionContext(properties),
			contributingFactors: getArrayProperty<TimelineEntryContributingFactor>(
				properties,
				"contributingFactors"
			),
			evidence: getArrayProperty<TimelineEntryEvidence>(properties, "evidence"),
			systemContext,
		},
	};
};

export const timelineEntryProperties = (
	attributes: TimelineAnalysisEntryAttributes
): CreateSystemAnalysisEntryAttributes["properties"] => ({
	isKey: attributes.isKey,
	decisionContext: attributes.decisionContext,
	contributingFactors: attributes.contributingFactors,
	evidence: attributes.evidence,
});
