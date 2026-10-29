<script lang="ts">
	import SearchInput from "$components/forms/search-input/SearchInput.svelte";
	import FilterOptions, { type FilterOption } from "$components/forms/filter-options/FilterOptions.svelte";
	import type { IncidentFilters, IncidentStatusFilter } from "../../lib/filters";

	type Props = {
		filters: IncidentFilters;
		onchange: (values: Partial<IncidentFilters>) => void;
		statusOptions: FilterOption<IncidentStatusFilter>[];
		severityOptions: FilterOption<string>[];
	};

	let { filters, onchange, statusOptions, severityOptions }: Props = $props();
</script>

<SearchInput value={filters.search} oninput={(search) => onchange({ search })} />
<FilterOptions
	label="Status"
	options={statusOptions}
	value={filters.status}
	onchange={(status) => onchange({ status })}
/>
<FilterOptions
	label="Severity"
	options={severityOptions}
	value={filters.severityId}
	onchange={(severityId) => onchange({ severityId })}
/>
