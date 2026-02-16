<script lang="ts">
	import RosterSelectField from "$src/components/roster-select-field/RosterSelectField.svelte";
	import { mdiCalendarRange } from "@mdi/js";

	type Props = {
		rosterId?: string;
	};
	let { 
		rosterId = $bindable(),
	}: Props = $props();

	const onRosterSelected = (id?: string) => {
		rosterId = id;
	}
</script>

<div class="flex gap-2">
	<RosterSelectField onSelected={onRosterSelected} selectedId={rosterId} />

	<!-- <DateRangeField
		label="Date Range"
		periodTypes={[PeriodType.Day]}
		classes={{
			field: { root: "gap-0", container: "pl-0 flex items-center h-full", prepend: "[&>span]:mr-2" },
		}}
		icon={mdiCalendarRange}
		bind:value={dateRange}
	/> -->
</div>