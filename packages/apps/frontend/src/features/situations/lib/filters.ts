import { z } from "zod";
import type { ListSituationsData } from "$lib/api";

type SituationListQuery = NonNullable<ListSituationsData["query"]>;

export const situationStatusValues = ["raised", "watching", "muted", "closed"] as const;

export type SituationStatusFilter = (typeof situationStatusValues)[number];

export type SituationStatusOption = {
	value: SituationStatusFilter;
	label: string;
	emptyMessage: string;
	query: Pick<SituationListQuery, "stage" | "muted">;
};

/** Status options in display order; each is a stage and muted filter. */
export const situationStatusOptions: SituationStatusOption[] = [
	{
		value: "raised",
		label: "Raised",
		emptyMessage: "No raised situations.",
		query: { stage: ["raised"], muted: false },
	},
	{
		value: "watching",
		label: "Watching",
		emptyMessage: "Nothing is being watched.",
		query: { stage: ["candidate"], muted: false },
	},
	{
		value: "muted",
		label: "Muted",
		emptyMessage: "No muted situations.",
		query: { stage: ["candidate", "raised"], muted: true },
	},
	{
		value: "closed",
		label: "Closed",
		emptyMessage: "No closed situations.",
		query: { stage: ["closed"] },
	},
];

export const situationFilterSchema = z.object({
	search: z.string().default("").catch(""),
	status: z.enum(situationStatusValues).default("raised").catch("raised"),
});

export type SituationFilters = z.infer<typeof situationFilterSchema>;

export function situationStatusOption(value: SituationStatusFilter) {
	return situationStatusOptions.find((option) => option.value === value) ?? situationStatusOptions[0];
}

export function situationSearchFilter(search: string) {
	return search.trim() || undefined;
}

export function situationQueryFilters(filters: SituationFilters): SituationListQuery {
	return { ...situationStatusOption(filters.status).query, search: situationSearchFilter(filters.search) };
}
