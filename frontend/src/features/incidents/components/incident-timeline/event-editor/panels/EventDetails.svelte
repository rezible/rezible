<script lang="ts">
    import { EditorContent } from "svelte-tiptap";
    import { Field, ToggleGroup, ToggleOption, Tooltip, Icon, TextField, Header, DatePickerField } from "svelte-ux";
    import { mdiMagnify, mdiExclamation, mdiBook, mdiBrain, mdiCalendar } from "@mdi/js";
    import { onMount } from 'svelte';
    import type { TimelineEvent } from "$features/incidents/components/incident-timeline/types";
    import { createMentionEditor } from '$features/incidents/lib/editor.svelte';
    import DateTimePickerField from '$components/date-time-field/DateTimePickerField.svelte';
    import type { DateTimeAnchor } from "$lib/api";

	type Props = {
		eventType: TimelineEvent["type"];
	}
	let { eventType = $bindable() }: Props = $props();

	let dateAnchor = $state<DateTimeAnchor>({date: new Date(), time: "09:00:00", timezone: Intl.DateTimeFormat().resolvedOptions().timeZone});
	
	const eventTypeOptions = [
		{label: "Observation", value: "observation", icon: mdiMagnify, hint: "What was noticed or detected"},
		{label: "Action", value: "action", icon: mdiExclamation, hint: "Steps taken or changes made"},
		{label: "Decision", value: "decision", icon: mdiBook, hint: "Choices made and their rationale"},
		{label: "Context", value: "context", icon: mdiBrain, hint: "Background information or ongoing conditions"},
	];

	const descriptionEditor = createMentionEditor("", "cursor-text focus:outline-none min-h-20");

	onMount(() => {
		return () => {
			descriptionEditor.destroy();
		}
	})
</script>

<div class="flex flex-col gap-2 flex-1">
	<TextField label="Title" value="" />

	<Field label="Event Type">
		<ToggleGroup bind:value={eventType} variant="fill" inset class="w-full">
			{#each eventTypeOptions as opt}
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

	<DateTimePickerField label="Time" current={dateAnchor} onChange={v => {console.log(v)}} exactTime />

	<Field label="Description" classes={{input: "block"}}>
		<EditorContent editor={descriptionEditor} />
	</Field>
</div>