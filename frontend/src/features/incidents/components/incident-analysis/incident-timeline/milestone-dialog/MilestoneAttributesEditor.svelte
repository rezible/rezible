<script lang="ts">
	import { Field, Header, Icon, ToggleGroup, ToggleOption, Tooltip } from "svelte-ux";
	import { mdiMagnify, mdiExclamation, mdiBook, mdiBrain, mdiFlag } from "@mdi/js";
	import type { IncidentMilestoneAttributes } from "$lib/api";
	import DateTimePickerField from "$components/date-time-field/DateTimePickerField.svelte";
	import { milestoneAttributes } from "./milestoneAttributes.svelte";

	type MilestoneKindOption = {
		label: string;
		value: IncidentMilestoneAttributes["kind"];
		icon: string;
		hint: string;
	}
	const milestoneKindOptions: MilestoneKindOption[] = [
		{
			label: "Impact Start",
			value: "impact",
			icon: mdiMagnify,
			hint: "impact",
		},
		{
			label: "Detection",
			value: "detection",
			icon: mdiExclamation,
			hint: "first detection",
		},
		{
			label: "Response",
			value: "investigation",
			icon: mdiBook,
			hint: "response begins",
		},
		{
			label: "Mitigation",
			value: "mitigation",
			icon: mdiBrain,
			hint: "impact is mitigated",
		},
		{
			label: "Resolution",
			value: "resolution",
			icon: mdiBrain,
			hint: "impact is resolved",
		},
	];
</script>

<div class="flex flex-col min-h-0 max-h-full flex-1 gap-2 p-2">
	<Field label="Kind">
		<ToggleGroup bind:value={milestoneAttributes.kind} variant="fill" inset class="w-full">
			{#each milestoneKindOptions as opt}
				<ToggleOption value={opt.value}>
					<Tooltip title={opt.hint}>
						<span class="flex items-center justify-center gap-2 px-2">
							<Icon data={opt.icon} />
							{opt.label}
						</span>
					</Tooltip>
				</ToggleOption>
			{/each}
		</ToggleGroup>
	</Field>

	<DateTimePickerField 
		label="Time"
		current={milestoneAttributes.timestamp}
		onChange={ts => (milestoneAttributes.timestamp = ts)}
		exactTime
	/>
</div>
