import { page } from "$app/state";
import { investigationHref } from "$features/situations/lib/routes";
import { Context } from "runed";
import { useSituationController } from "../controller.svelte";
import { evidenceChanged, observationGroups, type SourceRecord } from "../model";

export class SituationBriefController {
	private situationController = useSituationController();

	private situation = $derived(this.situationController.situation);
	private situationId = $derived(this.situation?.id ?? "");
	attrs = $derived(this.situation?.attributes);

	initialGroupId = $derived(this.attrs?.observationGroups[0]?.id);

	observations = $derived(observationGroups(this.attrs?.observationGroups ?? []));
	private inspectedKey = $state<string>();
	inspectedRecord = $derived(
		this.observations
			.flatMap((group) => group.records)
			.find((record) => record.key === this.inspectedKey)
	);
	inspectedGroup = $state("");
	sourceSheetOpen = $state(false);

	preview = $derived(this.situationController.investigation);
	evidenceChanged = $derived(evidenceChanged(
		this.attrs?.evidenceRevision ?? 0,
		this.attrs?.investigation?.completedRevision
	));
	previewChanged = $derived(
		!!this.situationController.investigationReport && this.evidenceChanged
	);

	previewHref = $derived(investigationHref(this.situationId, page.url.search));

	groupChoices = $state<Record<string, boolean>>({});
	groupOpen = (id: string) => this.groupChoices[id] ?? id === this.initialGroupId;
	setGroupOpen = (id: string, open: boolean) => (this.groupChoices[id] = open);
	setAllGroups = (open: boolean) => {
		for (const group of this.observations) this.groupChoices[group.id] = open;
	};

	inspectSource = (record: SourceRecord, groupTitle: string) => {
		this.inspectedKey = record.key;
		this.inspectedGroup = groupTitle;
		this.sourceSheetOpen = true;
	};
}

const ctx = new Context<SituationBriefController>("SituationBriefController");
export const initSituationBriefController = () => ctx.set(new SituationBriefController());
export const useSituationBriefController = () => ctx.get();
