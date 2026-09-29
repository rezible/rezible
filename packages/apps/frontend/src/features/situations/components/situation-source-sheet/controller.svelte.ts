import { createQuery } from "@tanstack/svelte-query";
import { getKnowledgeGraphEvidenceOptions } from "$lib/api";
import { Context, type Getter } from "runed";
import { isDefinitiveUnavailableError, type SourceTarget } from "../../views/situation/model";

export class SituationSourceSheetController {
	private getTarget: Getter<SourceTarget | undefined> = () => undefined;
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

	constructor(getTarget: Getter<SourceTarget | undefined>) {
		this.getTarget = getTarget;
	}
}

const ctx = new Context<SituationSourceSheetController>("SituationSourceSheetController");
export const initSituationSourceSheetController = (getTarget: Getter<SourceTarget | undefined>) =>
	ctx.set(new SituationSourceSheetController(getTarget));
export const useSituationSourceSheetController = () => ctx.get();
