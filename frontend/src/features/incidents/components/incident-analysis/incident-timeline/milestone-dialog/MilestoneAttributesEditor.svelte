<script lang="ts">
	import { Field, Icon, Tooltip } from "svelte-ux";
	import { mdiAlertDecagram, mdiAccountAlert, mdiAccountEye, mdiFireExtinguisher, mdiTimelineClock, mdiShape } from "@mdi/js";
	import { EditorContent } from "svelte-tiptap";
	import type { IncidentMilestoneAttributes } from "$lib/api";
	import ToggleGroup from "$components/toggle-group/ToggleGroup.svelte";
	import ToggleOption from "$components/toggle-group/ToggleOption.svelte";
	import DateTimePickerField from "$components/date-time-field/DateTimePickerField.svelte";
	import { milestoneAttributes } from "./milestoneAttributes.svelte";
	import { onMount } from "svelte";
	import { SvelteSet } from "svelte/reactivity";
	import { timeline } from "../timeline.svelte";

	type MilestoneKindOption = {
		label: string;
		value: IncidentMilestoneAttributes["kind"];
		icon: string;
		hint: string;
		unique?: boolean;
	};
	const milestoneKindOptions: MilestoneKindOption[] = [
		{
			label: "Impact Start",
			value: "impact",
			icon: mdiAlertDecagram,
			hint: "Impact begins",
			unique: true,
		},
		{
			label: "Detection",
			value: "detection",
			icon: mdiAccountAlert,
			hint: "Impact is detected (monitoring, alerts, user reports, etc)",
		},
		{
			label: "Response",
			value: "investigation",
			icon: mdiAccountEye,
			hint: "A human is investigating",
		},
		{
			label: "Mitigation",
			value: "mitigation",
			icon: mdiFireExtinguisher,
			hint: "Impact is mitigated",
		},
		{
			label: "Resolution",
			value: "resolution",
			icon: mdiTimelineClock,
			hint: "Impact is resolved",
		},
	];

	const seenKinds = $derived(new SvelteSet(timeline.milestones.map(m => m.attributes.kind)));

	onMount(milestoneAttributes.mountDescriptionEditor);
</script>

<div class="flex flex-col min-h-0 max-h-full flex-1 gap-2 p-2">
	<Field label="Kind" icon={mdiShape}>
		<ToggleGroup bind:value={milestoneAttributes.kind} variant="fill" inset class="w-full">
			{#each milestoneKindOptions as opt}
				<ToggleOption value={opt.value} disabled={opt.unique || seenKinds.has(opt.value)}>
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

	<Field label="Description" classes={{ root: "grow", container: "h-full", input: "block" }}>
		{#if milestoneAttributes.descriptionEditor}
			<EditorContent editor={milestoneAttributes.descriptionEditor} />
		{/if}
	</Field>
</div>
