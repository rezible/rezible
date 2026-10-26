import { reviewPanelModel } from "$features/incidents/lib/review";
import { Context } from "runed";
import { createQueries } from "@tanstack/svelte-query";
import { resolve } from "$app/paths";
import {
	getSituationOptions,
	type IncidentAttributes,
	type IncidentMilestone,
	type IncidentMilestoneAttributes,
} from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { formatDuration, isValidTime } from "$lib/time";
import { safeExternalUrl } from "$lib/utils";
import { situationStageStatus } from "$features/situations/lib/status";
import { createInvestigationConclusions } from "$features/situations/lib/conclusions.svelte";
import { incidentServiceImpacts } from "$features/incidents/lib/impacts";
import { investigationHref } from "$features/situations/lib/routes";
import { incidentSeverityStatus, milestoneLabels, retrospectiveStatus } from "$features/incidents/lib/status";
import { failedMitigationIds, orderedMilestones } from "$features/incidents/lib/response";
import { useIncidentView } from "../controller.svelte";

export type IncidentFact =
	| { key: string; label: string; kind: "status"; status: StatusPresentation }
	| { key: string; label: string; kind: "time"; value: string | undefined; fallback: string }
	| { key: string; label: string; kind: "text"; text: string };

export type IncidentReviewState =
	| { kind: "loading" }
	| { kind: "error"; retry: () => void }
	| {
			kind: "ready";
			status: StatusPresentation;
			explanation: string;
			ownerName?: string;
			reportHref: string | undefined;
			primaryAction: boolean;
	  };

export type IncidentMilestoneView = {
	id: string;
	label: string;
	marker: string;
	at: string;
	description: string | undefined;
	meta: string | undefined;
	/** Set on an impact recorded while mitigated. */
	caption: string | undefined;
	milestone: IncidentMilestone;
};

export type LinkedSituationView = {
	id: string;
	title: string;
	href: string;
	status: StatusPresentation;
	closedAt: string | undefined;
	/** The latest investigation conclusion, when one exists. */
	conclusion: string | undefined;
	reportHref: string | undefined;
};

// Conclusions are loaded for the first few linked situations only.
const CONCLUSION_LIMIT = 3;

const milestoneMarkers = {
	impact: "bg-status-danger-foreground",
	detection: "bg-border",
	investigation: "bg-border",
	mitigation: "bg-status-warning-foreground",
	resolution: "bg-status-neutral-foreground",
} satisfies Record<IncidentMilestoneAttributes["kind"], string>;

class IncidentOverviewController {
	view = useIncidentView();
	attributes = $derived(this.view.incident?.attributes);

	title = $derived(this.attributes?.title ?? "Incident");
	summary = $derived(this.attributes?.summary.trim() ?? "");

	facts = $derived(this.attributes ? this.buildFacts(this.attributes) : []);
	services = $derived(this.attributes ? incidentServiceImpacts(this.attributes) : []);

	private conclusionSituations = $derived((this.attributes?.situations ?? []).slice(0, CONCLUSION_LIMIT));
	private situationQueries = createQueries(() => ({
		queries: this.conclusionSituations.map((situation) => ({
			...getSituationOptions({ path: { id: situation.id } }),
			retry: false,
		})),
	}));
	private investigationIds = $derived(
		this.situationQueries.map((query) => query.data?.data.attributes.investigation?.investigation.id)
	);
	private conclusions = createInvestigationConclusions(() => this.investigationIds);

	milestones = $derived(this.buildMilestones());
	situationLinks = $derived(this.buildSituationLinks());
	declaredBy = $derived(this.attributes?.declaredBy?.attributes?.name);
	review = $derived(this.buildReview());
	teams = $derived(
		(this.attributes?.teams ?? []).map((link) => ({ id: link.team.id, name: link.team.attributes.name }))
	);
	tags = $derived(
		(this.attributes?.tags ?? []).map((tag) => ({
			id: tag.id,
			label: `${tag.attributes.key}: ${tag.attributes.value}`,
		}))
	);
	ticket = $derived(this.buildTicket());

	private buildFacts(attributes: IncidentAttributes): IncidentFact[] {
		let duration = formatDuration(attributes.openedAt, attributes.resolvedAt);
		if (duration && !attributes.resolvedAt) {
			duration = `${duration} so far`;
		}

		const facts: IncidentFact[] = [
			{
				key: "severity",
				label: "Severity",
				kind: "status",
				status: incidentSeverityStatus(attributes.severity),
			},
			{
				key: "type",
				label: "Type",
				kind: "text",
				text: attributes.type?.attributes?.name ?? "Unspecified",
			},
			{
				key: "opened",
				label: "Opened",
				kind: "time",
				value: this.validTime(attributes.openedAt),
				fallback: "Not recorded",
			},
			{
				key: "resolved",
				label: "Resolved",
				kind: "time",
				value: this.validTime(attributes.resolvedAt),
				fallback: "Not resolved",
			},
			{ key: "duration", label: "Duration", kind: "text", text: duration ?? "Unknown" },
		];

		for (const field of attributes.fieldSelections) {
			facts.push({
				key: `field:${field.fieldId}`,
				label: field.fieldName,
				kind: "text",
				text: field.option.attributes.value,
			});
		}
		return facts;
	}

	private validTime(value: string | null | undefined) {
		if (isValidTime(value)) {
			return value;
		}
		return undefined;
	}

	private buildMilestones(): IncidentMilestoneView[] {
		const milestones = this.attributes?.milestones ?? [];
		const failed = failedMitigationIds(milestones);
		return orderedMilestones(milestones).map((milestone) => {
			const kind = milestone.attributes.kind;
			// The API returns null for milestones without a user, despite the generated type.
			const userName = milestone.attributes.user?.attributes?.name;
			const source = milestone.attributes.source;
			let meta: string | undefined;
			if (userName) {
				meta = `by ${userName}`;
			} else if (source) {
				meta = `from ${source.charAt(0).toUpperCase()}${source.slice(1)}`;
			}
			let caption: string | undefined;
			if (failed.has(milestone.id)) {
				caption = "Mitigation failed";
			}
			return {
				id: milestone.id,
				label: milestoneLabels[kind] ?? "Milestone",
				marker: milestoneMarkers[kind] ?? "bg-border",
				at: milestone.attributes.timestamp,
				description: milestone.attributes.description.trim() || undefined,
				meta,
				caption,
				milestone,
			};
		});
	}

	private buildSituationLinks(): LinkedSituationView[] {
		return (this.attributes?.situations ?? []).map((situation, index) => {
			let conclusion: string | undefined;
			let reportHref: string | undefined;
			if (index < CONCLUSION_LIMIT) {
				const situationConclusion = this.conclusions.conclusionFor(this.investigationIds[index]);
				const situationReference = this.situationQueries[index]?.data?.data.attributes.reference;
				if (situationConclusion.kind === "text" && situationReference) {
					conclusion = situationConclusion.text;
					reportHref = investigationHref(situationReference);
				}
			}

			return {
				id: situation.id,
				title: situation.title,
				href: resolve("/situations/[ref]/[[view=situationView]]", { ref: situation.id }),
				status: situationStageStatus(situation.stage),
				closedAt: situation.closedAt,
				conclusion,
				reportHref,
			};
		});
	}

	private buildReview(): IncidentReviewState {
		const view = this.view;
		const incidentReference = this.attributes?.reference;

		if (!view.incidentRetrospectiveId) {
			return {
				kind: "ready",
				status: retrospectiveStatus(undefined),
				explanation: reviewPanelModel().explanation,
				reportHref: undefined,
				primaryAction: false,
			};
		}

		const retrospective = view.retrospective;
		if (!retrospective) {
			if (view.retrospectiveQuery.isError) {
				return {
					kind: "error",
					retry: () => {
						view.retrospectiveQuery.refetch();
					},
				};
			}
			return { kind: "loading" };
		}

		const state = retrospective.attributes.state;
		let reportHref: string | undefined;
		if (incidentReference) {
			reportHref = resolve("/incidents/[ref]/[[view=incidentView]]", {
				ref: incidentReference,
				view: "report",
			});
		}

		const reviewModel = reviewPanelModel(retrospective.attributes);
		return {
			kind: "ready",
			status: reviewModel.status,
			explanation: reviewModel.explanation,
			ownerName: retrospective.attributes.owner?.attributes?.name,
			reportHref,
			primaryAction: state !== "closed",
		};
	}

	private buildTicket() {
		const ticket = this.attributes?.externalTicket;
		const url = safeExternalUrl(ticket?.url);
		if (!url) {
			return undefined;
		}
		return { url, label: ticket?.title || "External ticket" };
	}
}

const context = new Context<IncidentOverviewController>("IncidentOverviewController");
export const initIncidentOverviewController = () => context.set(new IncidentOverviewController());
export const useIncidentOverviewController = () => context.get();
