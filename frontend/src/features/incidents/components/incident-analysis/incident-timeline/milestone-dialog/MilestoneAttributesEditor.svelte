<script lang="ts">
	import { Header } from "svelte-ux";
	import DateTimePickerField from "$components/date-time-field/DateTimePickerField.svelte";

	import { milestoneAttributes } from "./milestoneAttributes.svelte";
</script>

<div class="flex flex-row min-h-0 max-h-full flex-1 gap-2 p-2">
	<div class="flex flex-col w-1/2 gap-2 pl-1">
		<Header title="Details" />

		<DateTimePickerField
			label="Time"
			current={milestoneAttributes.timestamp}
			onChange={ts => (milestoneAttributes.timestamp = ts)}
			exactTime
		/>
	</div>

	
	<div class="flex flex-col gap-2 overflow-y-auto flex-1">
		<span>preview</span>
	</div>
</div>
