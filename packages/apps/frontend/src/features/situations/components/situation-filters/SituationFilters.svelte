<script lang="ts">
	import SearchInput from "$components/forms/search-input/SearchInput.svelte";
	import FilterOptions, { type FilterOption } from "$components/forms/filter-options/FilterOptions.svelte";
	import type { SituationFilters, SituationStatusFilter } from "../../lib/filters";

	type Props = {
		filters: SituationFilters;
		onchange: (values: Partial<SituationFilters>) => void;
		statusOptions: FilterOption<SituationStatusFilter>[];
	};

	let { filters, onchange, statusOptions }: Props = $props();
</script>

<SearchInput value={filters.search} oninput={(search) => onchange({ search })} />
<FilterOptions
	label="Status"
	options={statusOptions}
	value={filters.status}
	onchange={(status) => onchange({ status })}
/>
