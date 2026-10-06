import { z } from "zod";
import type { ListSituationsData } from "$lib/api";

type SituationListQuery = NonNullable<ListSituationsData["query"]>;

export const situationTabValues = ["raised", "watching", "muted", "closed"] as const;

export type SituationTab = (typeof situationTabValues)[number];

export type SituationTabOption = {
	value: SituationTab;
	label: string;
	emptyMessage: string;
	query: Pick<SituationListQuery, "stage" | "muted">;
};

/** List tabs in display order; each is a stage and muted filter. */
export const situationTabs: SituationTabOption[] = [
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
	tab: z.enum(situationTabValues).default("raised").catch("raised"),
});

export type SituationFilters = z.infer<typeof situationFilterSchema>;

export function situationTab(value: SituationTab) {
	return situationTabs.find((tab) => tab.value === value) ?? situationTabs[0];
}

export function situationSearchFilter(search: string) {
	return search.trim() || undefined;
}

export function situationQueryFilters(filters: SituationFilters): SituationListQuery {
	return { ...situationTab(filters.tab).query, search: situationSearchFilter(filters.search) };
}
