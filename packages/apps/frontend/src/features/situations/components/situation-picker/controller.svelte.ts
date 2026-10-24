import { Debounced, type Getter } from "runed";
import { listSituationsOptions, type Situation, type SituationAttributes } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { situationSearchFilter } from "$features/situations/lib/filters";
import { situationStatus } from "$features/situations/lib/status";

const PICKER_PAGE_SIZE = 10;
const SEARCH_DEBOUNCE_MS = 250;

export type SituationPickerOption = {
	id: string;
	title: string;
	status: ReturnType<typeof situationStatus>;
	openedAt: string;
	services: string[];
	situation: Situation;
};

/** Which situations a picker offers. */
export type SituationPickerScope = {
	stages: SituationAttributes["stage"][];
	/** Leave unset to offer muted and unmuted situations. */
	muted?: boolean;
};

const raisedAndWatching: SituationPickerScope = { stages: ["raised", "candidate"], muted: false };

/** Situations in the scope matching a search, other than the excluded one. By default, raised and watching ones. */
export class SituationPickerController {
	search = $state("");
	private debouncedSearch = new Debounced(() => this.search, SEARCH_DEBOUNCE_MS);
	private getExcludedId: Getter<string | undefined>;
	private scope = $state(raisedAndWatching);

	private preview = createPaginatedQuery({
		source: "local",
		queryOptions: () =>
			listSituationsOptions({
				query: {
					page: 1,
					pageSize: PICKER_PAGE_SIZE,
					stage: this.scope.stages,
					muted: this.scope.muted,
					search: situationSearchFilter(this.debouncedSearch.current),
				},
			}),
	});

	private query = $derived(this.preview.query);

	isLoading = $derived(this.query.isPending && !this.query.data);
	isError = $derived(this.query.isError && !this.query.data);
	options = $derived(this.getOptions());

	constructor(getExcludedId: Getter<string | undefined>, scope = raisedAndWatching) {
		this.getExcludedId = getExcludedId;
		this.scope = scope;
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
			const services = situation.attributes.entities
				.filter((entity) => entity.kind === "service")
				.map((entity) => entity.displayName);
			options.push({
				id: situation.id,
				title: situation.attributes.title,
				status: situationStatus(situation.attributes),
				openedAt: situation.attributes.openedAt,
				services,
				situation,
			});
		}
		return options;
	}
}
