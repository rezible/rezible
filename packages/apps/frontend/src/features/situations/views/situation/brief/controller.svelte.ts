import { page } from "$app/state";
import { investigationHref, situationHref } from "$features/situations/lib/routes";
import { Context } from "runed";
import { useSituationController } from "../controller.svelte";
import { evidenceChanged, latestInvestigation, observationGroups, type SourceRecord } from "../model";

export class SituationBriefController {
	private situationController = useSituationController();

	private situation = $derived(this.situationController.situation);
	private situationId = $derived(this.situation?.id ?? "");
	attrs = $derived(this.situation?.attributes);

	initialGroupId = $derived(this.attrs?.observationGroups[0]?.id);

	observations = $derived(observationGroups(this.attrs?.observationGroups ?? []));
	private inspectedKey = $state<string>();
	inspectedRecord = $derived(
		this.observations.items
			.flatMap((group) => group.records)
			.find((record) => record.key === this.inspectedKey)
	);
	inspectedGroup = $state("");
	sourceSheetOpen = $state(false);
	private sourceTrigger?: HTMLElement;
	private fallbackTarget?: HTMLElement;

	preview = $derived(latestInvestigation(this.attrs?.investigations ?? [], true));
	previewChanged = $derived(evidenceChanged(this.attrs?.evidenceRevision ?? 0, this.preview));

	previewHref = $derived(
		this.preview
			? investigationHref(this.situationId, this.preview.id, page.url.search)
			: situationHref(this.situationId, "investigations", page.url.search)
	);

	groupChoices = $state<Record<string, boolean>>({});
	groupOpen = (id: string) => this.groupChoices[id] ?? id === this.initialGroupId;
	setGroupOpen = (id: string, open: boolean) => (this.groupChoices[id] = open);
	setAllGroups = (open: boolean) => {
		for (const group of this.observations.items) this.groupChoices[group.id] = open;
	};

	inspectSource = (record: SourceRecord, groupTitle: string, trigger: HTMLElement) => {
		this.inspectedKey = record.key;
		this.inspectedGroup = groupTitle;
		this.sourceTrigger = trigger;
		this.sourceSheetOpen = true;
	};
	setFallbackTarget = (target: HTMLElement | undefined) => {
		this.fallbackTarget = target;
	};

	restoreSourceFocus = (event: Event) => {
		const target = this.sourceTrigger?.isConnected
			? this.sourceTrigger
			: this.fallbackTarget?.isConnected
				? this.fallbackTarget
				: undefined;
		if (!target) return;
		event.preventDefault();
		target.focus({ preventScroll: true });
	};
}

const ctx = new Context<SituationBriefController>("SituationBriefController")
export const initSituationBriefController = () => ctx.set(new SituationBriefController());
export const useSituationBriefController = () => ctx.get();
