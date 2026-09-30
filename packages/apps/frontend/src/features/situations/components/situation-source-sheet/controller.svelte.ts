import { createQuery } from "@tanstack/svelte-query";
import { getKnowledgeGraphEvidenceOptions } from "$lib/api";
import { Context, type Getter } from "runed";
import { isDefinitiveUnavailableError, type SourceTarget } from "$features/situations/lib/model";
import { formatTime } from "$lib/time";

export class SituationSourceSheetController {
	private getTarget = $state<Getter<SourceTarget | undefined>>(() => undefined);
	target = $derived(this.getTarget());
	directRecord = $derived(this.target?.kind === "direct" ? this.target.record : undefined);
	evidenceId = $derived(this.target?.kind === "knowledgeEvidence" ? this.target.id : undefined);

	evidenceQuery = createQuery(() => ({
		...getKnowledgeGraphEvidenceOptions({ path: { id: this.evidenceId ?? "" } }),
		enabled: !!this.evidenceId,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	evidence = $derived(
		isDefinitiveUnavailableError(this.evidenceQuery.error) ? undefined : this.evidenceQuery.data?.data
	);
	evidenceUnavailable = $derived(isDefinitiveUnavailableError(this.evidenceQuery.error));

	evidenceAttributes = $derived(this.evidence?.attributes);
	evidenceTime = $derived(formatTime(this.evidenceAttributes?.effectiveAt));
	evidenceKindLabel = $derived(this.evidenceAttributes?.kind === "observed" ? "Observed" : "Deleted");
	title = $derived(this.getTitle());
	description = $derived(this.getDescription());

	constructor(getTarget: Getter<SourceTarget | undefined>) {
		this.getTarget = getTarget;
	}

	private getTitle() {
		if (this.directRecord?.title) {
			return this.directRecord.title;
		}
		const subjectName = this.evidenceAttributes?.subjectState.displayName;
		if (subjectName) {
			return subjectName;
		}
		if (this.target?.kind === "knowledgeEvidence") {
			return "Knowledge evidence";
		}
		return "Source record";
	}

	private getDescription() {
		if (this.target?.kind === "direct") {
			return `Source in observation group: ${this.target.observationGroupTitle}`;
		}
		return "Existing knowledge evidence record cited by the report.";
	}
}

const ctx = new Context<SituationSourceSheetController>("SituationSourceSheetController");
export const initSituationSourceSheetController = (getTarget: Getter<SourceTarget | undefined>) =>
	ctx.set(new SituationSourceSheetController(getTarget));
export const useSituationSourceSheetController = () => ctx.get();
