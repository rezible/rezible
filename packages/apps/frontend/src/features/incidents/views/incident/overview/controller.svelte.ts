import { Context } from "runed";
import { createQueries } from "@tanstack/svelte-query";
import { resolve } from "$app/paths";
import { getSituationOptions, type IncidentAttributes, type IncidentMilestoneAttributes } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { formatDuration, isValidTime } from "$lib/time";
import { safeExternalUrl } from "$lib/utils";
import type { TimelineEntry } from "$components/common/timeline/Timeline.svelte";
import { situationStatus } from "$features/situations/lib/status";
import type { SituationConclusion } from "$features/situations/lib/model";
import { createInvestigationConclusions } from "$features/situations/lib/conclusions.svelte";
import { incidentServiceImpacts } from "$features/incidents/lib/impacts";
import { investigationHref } from "$features/situations/lib/routes";
import { incidentSeverityStatus, retrospectiveStatus } from "$features/incidents/lib/status";
import { useIncidentView } from "../controller.svelte";

export type IncidentFact =
	| { key: string; label: string; kind: "status"; status: StatusPresentation }
	| { key: string; label: string; kind: "time"; value: string | undefined; fallback: string }
	| { key: string; label: string; kind: "text"; text: string };

export type IncidentUnderstandingItem = {
	id: string;
	title: string;
	href: string;
	status: StatusPresentation | undefined;
	conclusion: SituationConclusion;
	reportHref: string | undefined;
};

export type IncidentReviewState =
	| { kind: "loading" }
	| { kind: "error"; retry: () => void }
	| {
			kind: "ready";
			status: StatusPresentation;
			explanation: string;
			reportHref: string | undefined;
			primaryAction: boolean;
			canStart: boolean;
	  };

export type IncidentRoleView = {
	id: string;
	userId: string | undefined;
	userName: string;
	roleName: string;
};

const UNDERSTANDING_LIMIT = 3;

const milestoneLabels = {
	impact: "Impact started",
	detection: "Detected",
	investigation: "Investigation began",
	mitigation: "Mitigated",
	resolution: "Resolved",
} satisfies Record<IncidentMilestoneAttributes["kind"], string>;

class IncidentOverviewController {
	view = useIncidentView();
	attributes = $derived(this.view.incident?.attributes);

	title = $derived(this.attributes?.title ?? "Incident");
	summary = $derived(this.attributes?.summary.trim() ?? "");

	facts = $derived(this.attributes ? this.buildFacts(this.attributes) : []);
	services = $derived(this.attributes ? incidentServiceImpacts(this.attributes) : []);

	private linkedSituations = $derived((this.attributes?.situations ?? []).slice(0, UNDERSTANDING_LIMIT));
	moreLinkedSituations = $derived(
		Math.max(0, (this.attributes?.situations.length ?? 0) - UNDERSTANDING_LIMIT)
	);
	private situationQueries = createQueries(() => ({
		queries: this.linkedSituations.map((situation) => ({
			...getSituationOptions({ path: { id: situation.id } }),
			retry: false,
		})),
	}));
	private investigationIds = $derived(
		this.situationQueries.map((query) => query.data?.data.attributes.investigation?.investigation.id)
	);
	private conclusions = createInvestigationConclusions(() => this.investigationIds);
	understanding = $derived(this.buildUnderstanding());

	milestones = $derived(this.buildMilestones());
	review = $derived(this.buildReview());
	roles = $derived(this.buildRoles());
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

	startRetrospective = () => {
		this.view.startRetrospective();
	};

	get startRetrospectivePending() {
		return this.view.startRetrospectiveMutation.isPending;
	}

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

	private buildUnderstanding(): IncidentUnderstandingItem[] {
		return this.linkedSituations.map((situation, index) => {
			const situationQuery = this.situationQueries[index];
			const loaded = situationQuery?.data?.data;

			let conclusion: SituationConclusion = { kind: "loading" };
			let status: StatusPresentation | undefined;
			if (loaded) {
				conclusion = this.conclusions.conclusionFor(this.investigationIds[index]);
				status = situationStatus(loaded.attributes);
			} else if (!situationQuery?.isPending) {
				conclusion = { kind: "none" };
			}

			let reportHref: string | undefined;
			if (conclusion.kind === "text") {
				reportHref = investigationHref(situation.id);
			}

			return {
				id: situation.id,
				title: situation.title,
				href: resolve("/situations/[id]/[[view=situationView]]", { id: situation.id }),
				status,
				conclusion,
				reportHref,
			};
		});
	}

	private buildMilestones(): TimelineEntry[] {
		const milestones = [...(this.attributes?.milestones ?? [])].sort(
			(left, right) => Date.parse(left.attributes.timestamp) - Date.parse(right.attributes.timestamp)
		);
		return milestones.map((milestone) => {
			// The API returns null for milestones without a user, despite the generated type.
			const userName = milestone.attributes.user?.attributes?.name;
			let meta: string | undefined;
			if (userName) {
				meta = `by ${userName}`;
			}
			return {
				key: milestone.id,
				label: milestoneLabels[milestone.attributes.kind] ?? "Milestone",
				at: milestone.attributes.timestamp,
				description: milestone.attributes.description.trim() || undefined,
				meta,
			};
		});
	}

	private buildReview(): IncidentReviewState {
		const view = this.view;
		const incidentSlug = this.attributes?.slug;

		if (!view.incidentRetrospectiveId) {
			let explanation = "A retrospective is created when the response ends.";
			if (view.canStartRetrospective) {
				explanation = "The response has ended. Start the retrospective to begin the review.";
			}
			return {
				kind: "ready",
				status: retrospectiveStatus(undefined),
				explanation,
				reportHref: undefined,
				primaryAction: false,
				canStart: view.canStartRetrospective,
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
		if (incidentSlug) {
			reportHref = resolve("/incidents/[slug]/[[view=incidentView]]", {
				slug: incidentSlug,
				view: "report",
			});
		}

		return {
			kind: "ready",
			status: retrospectiveStatus(state),
			explanation: this.reviewExplanation(state),
			reportHref,
			primaryAction: state !== "closed",
			canStart: false,
		};
	}

	private reviewExplanation(state: string) {
		switch (state) {
			case "draft":
				return "The team is writing the retrospective.";
			case "in_review":
				return "The retrospective is in review.";
			case "meeting":
				return "The review meeting is being held.";
			case "closed":
				return "The review is closed. The account is frozen.";
			default:
				return "The review stage is unknown.";
		}
	}

	private buildRoles(): IncidentRoleView[] {
		return (this.attributes?.roles ?? []).map(({ id, attributes }) => ({
			id,
			userId: attributes.user.attributes ? attributes.user.id : undefined,
			userName: attributes.user.attributes?.name ?? "Unknown user",
			roleName: attributes.role.attributes?.name ?? "Unknown role",
		}));
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
