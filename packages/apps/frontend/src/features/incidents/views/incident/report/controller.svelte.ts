import { Context } from "runed";
import type { RetrospectiveReportSection } from "$lib/api";
import { useIncidentView } from "../controller.svelte";

const sectionPrompts: Record<string, string> = {
	summary: "Summarise what happened and how it was resolved…",
	"customer-impact": "Describe what customers saw and for how long…",
	background: "Explain the context a reader needs to follow this incident…",
	lessons: "Note what the team learned and what should change…",
};

/** The muted prompt shown in an empty report section that the reader can edit. */
export function reportSectionPrompt(section: RetrospectiveReportSection) {
	const description = section.description.trim();
	if (description) {
		return description;
	}
	return sectionPrompts[section.field] ?? `Write the ${section.title.toLowerCase()}…`;
}

export const emptyReadOnlySectionText = "Nothing written yet.";

export class IncidentReportController {
	viewController = useIncidentView();
	retrospectiveId = $derived(this.viewController.retrospectiveId);
	retrospective = $derived(this.viewController.retrospective);
	sections = $derived(this.retrospective?.attributes.reportSections ?? []);

	documentAccess = $derived(this.viewController.documentAccess);
	canView = $derived(this.documentAccess?.canView ?? false);
	canEdit = $derived(this.documentAccess?.canEdit ?? false);

	editMode = $state(false);
	toggleEdit = () => {
		if (this.canEdit) this.editMode = !this.editMode;
	};
}

const ctx = new Context<IncidentReportController>("IncidentReportController");
export const initIncidentReportController = () => ctx.set(new IncidentReportController());
export const useIncidentReportController = () => ctx.get();
