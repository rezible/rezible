import { Context, watch } from "runed";
import { useSituationController } from "../controller.svelte";
import {
	observationGroups,
	timestamp,
	reportExcerpt,
	type SourceRecord,
} from "$features/situations/lib/model";
import { SourceInspection } from "$features/situations/lib/sourceInspection.svelte";

export class SituationOverviewController {
	private pageController = useSituationController();
	inspection = new SourceInspection();
	situationAttributes = $derived(this.pageController.situation?.attributes);
	observations = $derived(observationGroups(this.situationAttributes?.observationGroups ?? []));
	initialGroupId = $derived(this.situationAttributes?.observationGroups[0]?.id);
	groupChoices = $state<Record<string, boolean>>({});
	openedAt = $derived(timestamp(this.situationAttributes?.openedAt));
	closedAt = $derived(timestamp(this.situationAttributes?.closedAt));
	closeReason = $derived(this.getCloseReason());
	excerpt = $derived(this.getReportExcerpt());

	get investigationId() {
		return this.pageController.investigationId;
	}

	get investigationQuery() {
		return this.pageController.investigationQuery;
	}

	get investigationAttributes() {
		return this.pageController.investigationAttributes;
	}

	get investigationUnavailable() {
		return this.pageController.investigationUnavailable;
	}

	get execution() {
		return this.pageController.execution;
	}

	get reportQuery() {
		return this.pageController.reportQuery;
	}

	get reportAttributes() {
		return this.pageController.reportAttributes;
	}

	get reportPublishedAt() {
		return this.pageController.reportPublishedAt;
	}

	get reportAccessLost() {
		return this.pageController.reportAccessLost;
	}

	get investigationHref() {
		return this.pageController.investigationHref;
	}

	constructor() {
		watch(
			() => this.pageController.investigationId,
			() => this.inspection.reset(),
			{ lazy: true }
		);
	}

	groupOpen(id: string) {
		return this.groupChoices[id] ?? id === this.initialGroupId;
	}

	setGroupOpen = (id: string, open: boolean) => {
		this.groupChoices[id] = open;
	};
	inspectSource = (record: SourceRecord, groupTitle: string, trigger: HTMLElement) => {
		this.inspection.open({ kind: "direct", record, observationGroupTitle: groupTitle }, trigger);
	};

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

	private getReportExcerpt() {
		if (!this.reportAttributes) {
			return undefined;
		}
		return reportExcerpt(this.reportAttributes.text);
	}
}

const ctx = new Context<SituationOverviewController>("SituationOverviewController");
export const initSituationOverviewController = () => ctx.set(new SituationOverviewController());
export const useSituationOverviewController = () => ctx.get();
