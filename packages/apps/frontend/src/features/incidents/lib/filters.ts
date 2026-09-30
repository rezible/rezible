import { z } from "zod";
import type { ListIncidentsData } from "$lib/api";

export const incidentStatusOptions = [
	{ value: "active", label: "Active" },
	{ value: "any", label: "Any" },
	{ value: "unknown", label: "Unknown" },
	{ value: "started", label: "Started" },
	{ value: "mitigated", label: "Mitigated" },
	{ value: "resolved", label: "Resolved" },
] as const;

export const incidentFilterSchema = z.object({
	search: z.string().default("").catch(""),
	status: z
		.enum(["active", "any", "unknown", "started", "mitigated", "resolved"])
		.default("any")
		.catch("any"),
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
}: IncidentFilters): ListIncidentsData["query"] => ({
	search: search.trim() || undefined,
	responseStates: status === "any" ? undefined : status === "active" ? ["started", "mitigated"] : [status],
	severityId: severityId || undefined,
});
