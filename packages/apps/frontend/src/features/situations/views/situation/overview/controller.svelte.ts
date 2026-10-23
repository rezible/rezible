import { Context, watch } from "runed";
import { SvelteSet } from "svelte/reactivity";
import { createQueries } from "@tanstack/svelte-query";
import { getInvestigationProviderReadOptions, type ApiError } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { isValidTime } from "$lib/time";
import type { TimelineEntry } from "$components/common/timeline/Timeline.svelte";
import { useSituationController } from "../controller.svelte";
import {
	observationGroups,
	situationLinkKindLabel,
	situationSourceCount,
	type ReportState,
	type SourceRecord,
} from "$features/situations/lib/model";
import { situationHref } from "$features/situations/lib/routes";
import { closeReasonLabel, muteReasonLabel, situationStageStatus } from "$features/situations/lib/status";
import { SourceInspection } from "$features/situations/lib/sourceInspection.svelte";
import { whyPanelState } from "$features/situations/lib/why";
import {
	briefSections,
	providerReadResultView,
	type ProviderReadResultView,
} from "$features/situations/lib/brief";
import { isDefinitiveUnavailableError } from "$features/situations/lib/model";
import type { InvestigationOffer } from "../controller.svelte";
import { SituationNoticesController } from "./situationNotices.svelte";

export type UnderstandingState =
	| { kind: "loading" }
	| { kind: "none"; offer: InvestigationOffer }
	| { kind: "unavailable" }
	| { kind: "error" }
	| {
			kind: "investigation";
			run: StatusPresentation;
			/** The run notice shows the status itself, so the header shows it only without one. */
			showRunStatus: boolean;
			/** Shown when no report has been published. */
			noReportMessage: string;
			pendingFollowUp: boolean;
			report: ReportState;
			refreshFailed: boolean;
			href: string;
	  };

/** An expanded provider read: loading, failed to load, or what it returned. */
export type ReadExpansion =
	{ kind: "loading" } | { kind: "error"; retry: () => void } | ProviderReadResultView;

export type HeaderFact = { key: string; label: string; at: string; detail?: string };

export type RelatedSituation = {
	key: string;
	kindLabel: string;
	title: string;
	href: string;
	status: StatusPresentation;
};

export class SituationOverviewController {
	private pageController = useSituationController();

	inspection = new SourceInspection();
	notices = new SituationNoticesController(
		() => this.pageController.situation,
		(situation) => this.pageController.applySituation(situation)
	);

	situationAttributes = $derived(this.pageController.situation?.attributes);
	headerFacts = $derived(this.getHeaderFacts());
	why = $derived(
		this.situationAttributes
			? whyPanelState(this.situationAttributes, this.pageController.origin)
			: undefined
	);
	relatedSituations = $derived(this.getRelatedSituations());

	understanding = $derived(this.getUnderstanding());

	sections = $derived(briefSections(this.pageController.briefLoad, this.pageController.reportAttributes));

	expandedReadIds = new SvelteSet<string>();
	private expandedReads = $derived([...this.expandedReadIds]);
	private readQueries = createQueries(() => ({
		queries: this.expandedReads.map((readId) => ({
			...getInvestigationProviderReadOptions({
				path: { id: this.pageController.investigationId ?? "", readId },
			}),
			enabled: !!this.pageController.investigationId,
			retry: (failureCount: number, error: ApiError) =>
				!isDefinitiveUnavailableError(error) && failureCount < 2,
		})),
	}));
	readExpansions = $derived(this.getReadExpansions());

	observations = $derived(
		observationGroups(
			this.situationAttributes?.observationGroups ?? [],
			this.pageController.episodesByEntity
		)
	);
	sourceCount = $derived(this.situationAttributes ? situationSourceCount(this.situationAttributes) : 0);
	private initialGroupId = $derived(this.situationAttributes?.observationGroups[0]?.id);
	private groupChoices = $state<Record<string, boolean>>({});

	timeline = $derived(this.getTimeline());

	constructor() {
		watch(
			() => this.pageController.investigationId,
			() => {
				this.inspection.reset();
				this.expandedReadIds.clear();
			},
			{ lazy: true }
		);
	}

	get situationQuery() {
		return this.pageController.situationQuery;
	}

	get situationUnavailable() {
		return this.pageController.situationUnavailable;
	}

	get originLoading() {
		return this.pageController.originLoading;
	}

	get longRunning() {
		return this.pageController.longRunning;
	}

	get raisePending() {
		return this.pageController.raiseMutation.isPending;
	}

	get actionPending() {
		return this.pageController.actionPending;
	}

	get linkedIncidents() {
		return this.pageController.linkedIncidents;
	}

	get linkedIncidentsLoading() {
		return this.pageController.linkedIncidentsLoading;
	}

	get briefRefreshFailed() {
		return this.pageController.briefLoad.kind === "loaded" && this.pageController.briefQuery.isError;
	}

	raise = () => {
		this.pageController.raise();
	};

	unmute = () => {
		this.pageController.unmute();
	};

	retryInvestigation = () => {
		this.pageController.retryInvestigation();
	};

	retryReport = () => {
		this.pageController.retryReport();
	};

	retryBrief = () => {
		this.pageController.retryBrief();
	};

	toggleRead = (readId: string) => {
		if (this.expandedReadIds.has(readId)) {
			this.expandedReadIds.delete(readId);
			return;
		}
		this.expandedReadIds.add(readId);
	};

	/** Opens the evidence that reported a deployment with no link of its own. */
	openDeploymentEvidence = (evidenceId: string, trigger: HTMLElement) => {
		this.inspection.openEvidence(evidenceId, trigger);
	};

	groupOpen(id: string) {
		return this.groupChoices[id] ?? id === this.initialGroupId;
	}

	setGroupOpen = (id: string, open: boolean) => {
		this.groupChoices[id] = open;
	};

	expandAll = () => {
		this.setAllGroups(true);
	};

	collapseAll = () => {
		this.setAllGroups(false);
	};

	inspectSource = (record: SourceRecord, groupTitle: string, trigger: HTMLElement) => {
		this.inspection.open({ kind: "direct", record, observationGroupTitle: groupTitle }, trigger);
	};

	private getReadExpansions() {
		const expansions = new Map<string, ReadExpansion>();
		this.expandedReads.forEach((readId, index) => {
			const query = this.readQueries[index];
			const read = query?.data?.data;
			if (read) {
				expansions.set(readId, providerReadResultView(read));
			} else if (query?.isError) {
				expansions.set(readId, { kind: "error", retry: () => query.refetch() });
			} else {
				expansions.set(readId, { kind: "loading" });
			}
		});
		return expansions;
	}

	private setAllGroups(open: boolean) {
		const choices: Record<string, boolean> = {};
		for (const group of this.observations) {
			choices[group.id] = open;
		}
		this.groupChoices = choices;
	}

	private getHeaderFacts() {
		const attributes = this.situationAttributes;
		if (!attributes) {
			return [];
		}

		const facts: HeaderFact[] = [{ key: "first-signal", label: "First signal", at: attributes.openedAt }];
		facts.push({ key: "watching", label: "Watching since", at: attributes.createdAt });
		if (attributes.raisedAt) {
			facts.push({ key: "raised", label: "Raised", at: attributes.raisedAt });
		}
		if (attributes.mutedAt) {
			facts.push({
				key: "muted",
				label: "Muted",
				at: attributes.mutedAt,
				detail: muteReasonLabel(attributes.muteReason),
			});
		}
		if (this.pageController.holdUntil) {
			facts.push({ key: "held", label: "Held until", at: this.pageController.holdUntil });
		}
		if (attributes.closedAt) {
			facts.push({
				key: "closed",
				label: "Closed",
				at: attributes.closedAt,
				detail: closeReasonLabel(attributes.closeReason),
			});
		}
		return facts.filter((fact) => isValidTime(fact.at));
	}

	private getRelatedSituations() {
		const links = this.situationAttributes?.links ?? [];
		return links.map((link): RelatedSituation => ({
			key: `${link.kind}:${link.situationId}`,
			kindLabel: situationLinkKindLabel(link.kind),
			title: link.title,
			href: situationHref(link.situationId),
			status: situationStageStatus(link.stage),
		}));
	}

	private getUnderstanding(): UnderstandingState {
		const page = this.pageController;

		if (!this.situationAttributes) {
			return { kind: "loading" };
		}
		if (page.investigationOffer) {
			return { kind: "none", offer: page.investigationOffer };
		}
		if (page.investigationUnavailable) {
			return { kind: "unavailable" };
		}
		if (!page.investigationAttributes) {
			if (page.investigationQuery.isError) {
				return { kind: "error" };
			}
			return { kind: "loading" };
		}

		return {
			kind: "investigation",
			run: page.runStatus,
			showRunStatus: !page.updateNotice,
			noReportMessage: this.getNoReportMessage(),
			pendingFollowUp: page.investigationAttributes.hasPendingWork,
			report: page.reportState,
			refreshFailed: page.investigationQuery.isError,
			href: page.investigationHref,
		};
	}

	private getNoReportMessage() {
		const page = this.pageController;
		if (page.investigationAttributes?.activeTurn) {
			return "The investigation is running. Its report will appear here.";
		}
		if (page.runStatus.description) {
			return page.runStatus.description;
		}
		return "No report has been published yet.";
	}

	private getTimeline() {
		const attributes = this.situationAttributes;
		const investigation = this.pageController.investigationAttributes;
		const report = this.pageController.reportAttributes;
		const candidates: TimelineEntry[] = [];

		if (attributes) {
			candidates.push({ key: "opened", label: "First signal", at: attributes.openedAt });
		}
		if (attributes?.raisedAt) {
			candidates.push({ key: "raised", label: "Situation raised", at: attributes.raisedAt });
		}
		if (attributes?.mutedAt) {
			candidates.push({
				key: "muted",
				label: "Situation muted",
				at: attributes.mutedAt,
				description: muteReasonLabel(attributes.muteReason),
			});
		}
		if (investigation) {
			candidates.push({
				key: "investigation-started",
				label: "Investigation started",
				at: investigation.createdAt,
			});
		}
		if (report) {
			let description: string | undefined;
			if (report.provisional) {
				description = "Provisional";
			}
			candidates.push({
				key: "report-published",
				label: "Report published",
				at: report.createdAt,
				description,
			});
		}
		if (attributes?.closedAt) {
			candidates.push({
				key: "closed",
				label: "Situation closed",
				at: attributes.closedAt,
				description: closeReasonLabel(attributes.closeReason),
			});
		}

		const entries = candidates.filter((entry) => isValidTime(entry.at));
		entries.sort((first, second) => Date.parse(second.at) - Date.parse(first.at));
		return entries;
	}
}

const ctx = new Context<SituationOverviewController>("SituationOverviewController");
export const initSituationOverviewController = () => ctx.set(new SituationOverviewController());
export const useSituationOverviewController = () => ctx.get();
