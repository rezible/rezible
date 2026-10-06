<script lang="ts">
	import { Input } from "$components/ui/input";
	import type { SituationFilters } from "../../lib/filters";

	type Props = {
		filters: SituationFilters;
		onchange: (values: Partial<SituationFilters>) => void;
	};

	let { filters, onchange }: Props = $props();

	const id = $props.id();
</script>

<div class="min-w-0 basis-60">
	<label for={`${id}-search`} class="sr-only">Search situation titles</label>
	<Input
		id={`${id}-search`}
		type="search"
		placeholder="Search situation titles"
		value={filters.search}
		oninput={(event) => onchange({ search: event.currentTarget.value })}
	/>
</div>
