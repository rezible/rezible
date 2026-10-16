import type {
	KnowledgeGraphEntity,
	KnowledgeGraphEntityAlias,
	KnowledgeProposal,
	KnowledgeProposalDecision,
	KnowledgeProposalEvidence,
	KnowledgeProposalImplication,
	KnowledgeProposalValue,
} from "$lib/api";
import { formatTime } from "$lib/time";

export type ProposalAction = "confirm" | "reject";

export type ProposalStatusView = {
	label: string;
	tone: "pending" | "rejected" | "confirmed";
	actions: ProposalAction[];
};

export type DecisionView = {
	key: string;
	text: string;
	time: string;
	timeTitle: string;
};

export type AliasView = {
	key: string;
	integration: string;
	name: string;
	lastRecorded: string;
	lastRecordedTitle: string;
	/** Set when the alias joined the entity through a confirmed match. */
	mergedNote?: string;
};

export type ProposalValueView = {
	key: string;
	value: string;
	/** The entity's display name, when it differs from the value. */
	displayName?: string;
	aliases: AliasView[];
	/** No entity holds the value any more. */
	unheld: boolean;
};

export type EvidenceView = {
	integration: string;
	name: string;
	resourceRef: string;
	assertion: string;
	recorded: string;
	recordedTitle: string;
};

export type RationalePart = {
	key: string;
	text: string;
	/** Rezible writes values in backticks; they are shown as code. */
	code: boolean;
};

export type ProposalView = {
	id: string;
	/** "checkout and checkout-api". */
	label: string;
	status: ProposalStatusView;
	values: ProposalValueView[];
	/** Set when both values already belong to one entity, whose aliases are shown once. */
	sharedEntity?: { kind: string; aliases: AliasView[] };
	rationale: RationalePart[];
	origin: string;
	evidence: EvidenceView;
	decisions: DecisionView[];
	created: string;
	createdTitle: string;
};

export type ImpliedProposalView = {
	key: string;
	values: string;
	/** Who rejected it and when, for a rejected proposal. */
	rejection?: string;
};

export type ConfirmQuestion = {
	title: string;
	consequence: string;
};

export type EntityRowView = {
	id: string;
	displayName: string;
	values: string[];
	aliases: AliasView[];
	/** A pending proposal naming the entity. */
	possibleMatchProposalId?: string;
};

/** The label and the decisions a person can make for a proposal in its current status. */
export function proposalStatusView(status: KnowledgeProposal["attributes"]["status"]): ProposalStatusView {
	if (status === "pending") {
		return { label: "Proposed match", tone: "pending", actions: ["confirm", "reject"] };
	}
	if (status === "rejected") {
		return { label: "Rejected", tone: "rejected", actions: ["confirm"] };
	}
	return { label: "Confirmed", tone: "confirmed", actions: [] };
}

export function originLabel(origin: KnowledgeProposal["attributes"]["origin"]): string {
	if (origin === "identity_conflict") {
		return "A source names it differently";
	}
	return "Similar names";
}

export function decisionViews(decisions: KnowledgeProposalDecision[], now = Date.now()): DecisionView[] {
	return decisions.map((decision, index) => {
		const time = formatTime(decision.createdAt, now);
		let text: string;
		if (decision.status === "rejected") {
			text = `Rejected by ${decision.user.name}`;
		} else if (decision.impliedByProposalId) {
			text = `Confirmed by ${decision.user.name} with another match`;
		} else {
			text = `Confirmed by ${decision.user.name}`;
		}
		return { key: `${index}:${decision.createdAt}`, text, time: time.relative, timeTitle: time.absolute };
	});
}

export function aliasViews(aliases: KnowledgeGraphEntityAlias[], now = Date.now()): AliasView[] {
	return aliases.map((alias) => {
		const recorded = formatTime(alias.lastRecordedAt, now);
		const view: AliasView = {
			key: alias.id,
			integration:
				alias.integrationDisplayName || alias.provider || alias.attributes.resourceRef.provider,
			name: alias.name || alias.attributes.resourceRef.resourceRef,
			lastRecorded: `last recorded ${recorded.relative}`,
			lastRecordedTitle: recorded.absolute,
		};
		if (!alias.lastRecordedAt) {
			view.lastRecorded = "nothing recorded";
		}

		const merge = alias.mergedByProposal;
		if (merge) {
			const confirmedAt = formatTime(merge.confirmedAt, now);
			const confirmer = merge.confirmedBy?.name ?? "someone";
			view.mergedNote = `Joined through a match confirmed by ${confirmer}, ${confirmedAt.relative}`;
		}
		return view;
	});
}

export function proposalValueViews(values: KnowledgeProposalValue[], now = Date.now()): ProposalValueView[] {
	return values.map((value) => {
		const entity = value.entity;
		if (!entity) {
			return { key: value.value, value: value.value, aliases: [], unheld: true };
		}

		const view: ProposalValueView = {
			key: value.value,
			value: value.value,
			aliases: aliasViews(entity.aliases, now),
			unheld: false,
		};
		if (entity.displayName && entity.displayName !== value.value) {
			view.displayName = entity.displayName;
		}
		return view;
	});
}

export function evidenceView(evidence: KnowledgeProposalEvidence, now = Date.now()): EvidenceView {
	const recorded = formatTime(evidence.effectiveAt, now);
	return {
		integration: evidence.integrationDisplayName || evidence.provider,
		name: evidence.displayName || evidence.resourceRef,
		resourceRef: evidence.resourceRef,
		assertion: evidence.assertion.replaceAll("_", " "),
		recorded: recorded.relative,
		recordedTitle: recorded.absolute,
	};
}

export function rationaleParts(rationale: string): RationalePart[] {
	return rationale
		.split("`")
		.map((text, index) => ({ key: `${index}`, text, code: index % 2 === 1 }))
		.filter((part) => part.text !== "");
}

function sharedEntity(proposal: KnowledgeProposal, now: number): ProposalView["sharedEntity"] {
	const [first, second] = proposal.attributes.values.map((value) => value.entity);
	if (!first || !second || first.id !== second.id) {
		return undefined;
	}
	return { kind: proposal.attributes.kind, aliases: aliasViews(first.aliases, now) };
}

export function proposalView(proposal: KnowledgeProposal, now = Date.now()): ProposalView {
	const attributes = proposal.attributes;
	const created = formatTime(attributes.createdAt, now);
	return {
		id: proposal.id,
		label: attributes.values.map((value) => value.value).join(" and "),
		status: proposalStatusView(attributes.status),
		values: proposalValueViews(attributes.values, now),
		sharedEntity: sharedEntity(proposal, now),
		rationale: rationaleParts(attributes.rationale),
		origin: originLabel(attributes.origin),
		evidence: evidenceView(attributes.evidence, now),
		decisions: decisionViews(attributes.decisions, now),
		created: created.relative,
		createdTitle: created.absolute,
	};
}

/** What confirming asks before it merges: the two values, as one entity of the proposal's kind. */
export function confirmQuestion(proposal: KnowledgeProposal): ConfirmQuestion {
	const [first, second] = proposal.attributes.values.map((value) => value.value);
	return {
		title: `Treat ${first} and ${second} as one ${proposal.attributes.kind}?`,
		consequence: "Rezible will merge them. This can't be undone.",
	};
}

export function impliedProposalViews(
	implies: KnowledgeProposalImplication[],
	now = Date.now()
): ImpliedProposalView[] {
	return implies.map((implied) => {
		const view: ImpliedProposalView = {
			key: implied.id,
			values: implied.values.join(" and "),
		};
		if (implied.status === "rejected") {
			const rejecter = implied.rejectedBy?.name ?? "someone";
			const rejectedAt = formatTime(implied.rejectedAt, now);
			view.rejection = `rejected by ${rejecter}, ${rejectedAt.relative}`;
		}
		return view;
	});
}

/** The proposal naming each entity, by entity ID, from a list of pending proposals; the last listed wins. */
export function possibleMatchesByEntity(proposals: KnowledgeProposal[]): Map<string, string> {
	const matches = new Map<string, string>();
	for (const proposal of proposals) {
		for (const value of proposal.attributes.values) {
			const entityId = value.entity?.id;
			if (entityId) {
				matches.set(entityId, proposal.id);
			}
		}
	}
	return matches;
}

export function entityRowView(
	entity: KnowledgeGraphEntity,
	possibleMatches: Map<string, string>,
	now = Date.now()
): EntityRowView {
	const attributes = entity.attributes;
	const values = attributes.linkingAttributes.map((linkingAttribute) => linkingAttribute.value);
	let displayName = attributes.displayName;
	if (!displayName) {
		displayName = values[0] ?? "Unnamed";
	}
	return {
		id: entity.id,
		displayName,
		values,
		aliases: aliasViews(attributes.aliases, now),
		possibleMatchProposalId: possibleMatches.get(entity.id),
	};
}
