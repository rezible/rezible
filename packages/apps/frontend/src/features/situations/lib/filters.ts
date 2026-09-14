import { z } from "zod";

export const situationStatusOptions = [
	{ value: "any", label: "Any" },
	{ value: "active", label: "Active" },
	{ value: "investigating", label: "Investigating" },
	{ value: "closed", label: "Closed" },
] as const;

export const situationFilterSchema = z.object({
	search: z.string().default("").catch(""),
	status: z.enum(["any", "active", "investigating", "closed"]).default("any").catch("any"),
});

export type SituationFilters = z.infer<typeof situationFilterSchema>;

export const situationQueryFilters = (filters: SituationFilters) => ({
	search: filters.search.trim() || undefined,
	status: filters.status === "any" ? undefined : filters.status,
});
