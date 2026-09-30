import { Context, watch } from "runed";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { isValidTime } from "$lib/time";
import type { TimelineEntry } from "$components/common/timeline/Timeline.svelte";
import { useSituationController } from "../controller.svelte";
import { observationGroups, type ReportState, type SourceRecord } from "$features/situations/lib/model";
import { SourceInspection } from "$features/situations/lib/sourceInspection.svelte";

export type UnderstandingState =
	| { kind: "loading" }
	| { kind: "none" }
	| { kind: "unavailable" }
	| { kind: "error" }
	| {
			kind: "investigation";
			run: StatusPresentation;
			/** Shown when no report has been published. */
			noReportMessage: string;
			pendingFollowUp: boolean;
			report: ReportState;
			refreshFailed: boolean;
			href: string;
	  };

export class SituationOverviewController {
	private pageController = useSituationController();

	inspection = new SourceInspection();

	situationAttributes = $derived(this.pageController.situation?.attributes);
	openedAt = $derived(this.situationAttributes?.openedAt);
	closedAt = $derived(this.situationAttributes?.closedAt);
	closeReason = $derived(this.getCloseReason());

	understanding = $derived(this.getUnderstanding());

	observations = $derived(observationGroups(this.situationAttributes?.observationGroups ?? []));
	sourceCount = $derived(this.observations.reduce((total, group) => total + group.records.length, 0));
	private initialGroupId = $derived(this.situationAttributes?.observationGroups[0]?.id);
	private groupChoices = $state<Record<string, boolean>>({});

	timeline = $derived(this.getTimeline());

	constructor() {
		watch(
			() => this.pageController.investigationId,
			() => this.inspection.reset(),
			{ lazy: true }
		);
	}

	get situationQuery() {
		return this.pageController.situationQuery;
	}

	get situationUnavailable() {
		return this.pageController.situationUnavailable;
	}

	get startPending() {
		return this.pageController.requestInvestigationMutation.isPending;
	}

	get linkedIncidents() {
		return this.pageController.linkedIncidents;
	}

	get linkedIncidentsLoading() {
		return this.pageController.linkedIncidentsLoading;
	}

	startInvestigation = () => {
		this.pageController.startInvestigation();
	};

	retryInvestigation = () => {
		this.pageController.retryInvestigation();
	};

	retryReport = () => {
		this.pageController.retryReport();
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

	private setAllGroups(open: boolean) {
		const choices: Record<string, boolean> = {};
		for (const group of this.observations) {
			choices[group.id] = open;
		}
		this.groupChoices = choices;
	}

	private getCloseReason() {
		switch (this.situationAttributes?.closeReason) {
			case "stabilized":
				return "Stabilized";
			case "dismissed":
				return "Dismissed";
			default:
				return undefined;
		}
	}

	private getUnderstanding(): UnderstandingState {
		const page = this.pageController;

		if (!this.situationAttributes) {
			return { kind: "loading" };
		}
		if (!page.investigationId) {
			return { kind: "none" };
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
			candidates.push({ key: "opened", label: "Situation opened", at: attributes.openedAt });
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
				description: this.closeReason,
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
