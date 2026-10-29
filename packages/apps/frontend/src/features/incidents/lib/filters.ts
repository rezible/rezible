import { z } from "zod";
import type { ListIncidentsData } from "$lib/api";

export type IncidentStatusFilter = "any" | "active" | "resolved";

export const incidentStatusOptions: { value: IncidentStatusFilter; label: string }[] = [
	{ value: "any", label: "All" },
	{ value: "active", label: "Active" },
	{ value: "resolved", label: "Resolved" },
];

export const incidentFilterSchema = z.object({
	search: z.string().default("").catch(""),
	status: z.enum(["active", "resolved", "any"]).default("any").catch("any"),
	severityId: z
		.union([z.uuid(), z.literal("")])
		.default("")
		.catch(""),
});

export type IncidentFilters = z.infer<typeof incidentFilterSchema>;

const defaultIncidentFilters = incidentFilterSchema.parse({});

/** How many filters differ from their defaults. */
export function incidentActiveFilterCount(filters: IncidentFilters) {
	return [
		filters.search !== defaultIncidentFilters.search,
		filters.status !== defaultIncidentFilters.status,
		filters.severityId !== defaultIncidentFilters.severityId,
	].filter(Boolean).length;
}

export const incidentQueryFilters = ({
	search,
	status,
	severityId,
}: IncidentFilters): ListIncidentsData["query"] => {
	let resolved: boolean | undefined;
	if (status === "active") {
		resolved = false;
	} else if (status === "resolved") {
		resolved = true;
	}

	return {
		search: search.trim() || undefined,
		resolved,
		severityId: severityId || undefined,
	};
};
