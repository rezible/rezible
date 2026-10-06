import { Debounced, type Getter } from "runed";
import { listSituationsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { situationSearchFilter } from "$features/situations/lib/filters";
import { situationStatus } from "$features/situations/lib/status";

const PICKER_PAGE_SIZE = 10;
const SEARCH_DEBOUNCE_MS = 250;

export type SituationPickerOption = {
	id: string;
	title: string;
	status: ReturnType<typeof situationStatus>;
};

/** Raised and watching situations matching a search, other than the excluded one. */
export class SituationPickerController {
	search = $state("");
	private debouncedSearch = new Debounced(() => this.search, SEARCH_DEBOUNCE_MS);
	private getExcludedId: Getter<string | undefined>;

	private preview = createPaginatedQuery({
		source: "local",
		queryOptions: () =>
			listSituationsOptions({
				query: {
					page: 1,
					pageSize: PICKER_PAGE_SIZE,
					stage: ["raised", "candidate"],
					muted: false,
					search: situationSearchFilter(this.debouncedSearch.current),
				},
			}),
	});

	private query = $derived(this.preview.query);

	isLoading = $derived(this.query.isPending && !this.query.data);
	isError = $derived(this.query.isError && !this.query.data);
	options = $derived(this.getOptions());

	constructor(getExcludedId: Getter<string | undefined>) {
		this.getExcludedId = getExcludedId;
	}

	setSearch = (value: string) => {
		this.search = value;
	};

	retry = () => {
		this.query.refetch();
	};

	private getOptions(): SituationPickerOption[] {
		const excludedId = this.getExcludedId();
		const options: SituationPickerOption[] = [];
		for (const situation of this.query.data?.data ?? []) {
			if (situation.id === excludedId) {
				continue;
			}
			options.push({
				id: situation.id,
				title: situation.attributes.title,
				status: situationStatus(situation.attributes),
			});
		}
		return options;
	}
}
