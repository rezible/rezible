import { createQueries } from "@tanstack/svelte-query";
import type { Getter } from "runed";
import { getInvestigationReportOptions } from "$lib/api";
import { situationConclusion, type SituationConclusion } from "./model";

/**
 * Loads the latest report for each investigation ID and exposes the resulting
 * conclusions. Missing IDs (situations without an investigation) have no conclusion.
 */
export function createInvestigationConclusions(getInvestigationIds: Getter<(string | undefined)[]>) {
	const investigationIds = $derived([...new Set(getInvestigationIds().filter((id): id is string => !!id))]);

	const reportQueries = createQueries(() => ({
		queries: investigationIds.map((id) => ({
			...getInvestigationReportOptions({ path: { id } }),
			retry: false,
		})),
	}));

	const queriesByInvestigationId = $derived(
		new Map(investigationIds.map((id, index) => [id, reportQueries[index]]))
	);

	return {
		conclusionFor(investigationId: string | undefined): SituationConclusion {
			if (!investigationId) {
				return situationConclusion(undefined);
			}
			const query = queriesByInvestigationId.get(investigationId);
			if (!query) {
				return { kind: "loading" };
			}
			return situationConclusion(query);
		},
	};
}
