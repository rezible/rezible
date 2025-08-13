<script lang="ts">
	import RosterSelectField from "$src/components/roster-select-field/RosterSelectField.svelte";
	import { DateRangeField } from "svelte-ux";
	import { type DateRange as DateRangeType } from '@layerstack/utils/dateRange';
	import { PeriodType } from "@layerstack/utils";
	import { mdiCalendarRange } from "@mdi/js";

	type Props = {
		rosterId?: string;
		dateRange?: DateRangeType;
	};
	let { 
		rosterId = $bindable(),
		dateRange = $bindable(),
	}: Props = $props();

	const onRosterSelected = (id?: string) => {
		rosterId = id;
	}
</script>

<div class="flex gap-2">
	<RosterSelectField onSelected={onRosterSelected} selectedId={rosterId} classes={{ root: "w-64" }} />

	<DateRangeField
		label="Date Range"
		periodTypes={[PeriodType.Day]}
		classes={{
			field: { root: "gap-0", container: "pl-0 flex items-center h-full", prepend: "[&>span]:mr-2" },
		}}
		icon={mdiCalendarRange}
		bind:value={dateRange}
	/>
</div>