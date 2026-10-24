import { z } from "zod";
import type { ListIncidentsData } from "$lib/api";

export const incidentStatusOptions = [
	{ value: "active", label: "Active" },
	{ value: "resolved", label: "Resolved" },
	{ value: "any", label: "Any" },
] as const;

export const incidentFilterSchema = z.object({
	search: z.string().default("").catch(""),
	status: z.enum(["active", "resolved", "any"]).default("any").catch("any"),
	severityId: z
		.union([z.uuid(), z.literal("")])
		.default("")
		.catch(""),
});

export type IncidentFilters = z.infer<typeof incidentFilterSchema>;

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
