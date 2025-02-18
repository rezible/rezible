<script lang="ts">
	import { EditorContent } from "svelte-tiptap";
	import {
		Field,
		ToggleGroup,
		ToggleOption,
		Tooltip,
		Icon,
		TextField,
		Header,
		DatePickerField,
		Switch,
	} from "svelte-ux";
	import { mdiMagnify, mdiExclamation, mdiBook, mdiBrain, mdiCalendar, mdiFlag } from "@mdi/js";
	import { onMount } from "svelte";
	import { createMentionEditor } from "$features/incidents/lib/editor.svelte";
	import DateTimePickerField from "$components/date-time-field/DateTimePickerField.svelte";
	import type { DateTimeAnchor } from "$lib/api";
	import { eventAttributes } from "./eventAttributes.svelte";

	let dateAnchor = $state<DateTimeAnchor>({
		date: new Date(),
		time: "09:00:00",
		timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
	});

	const eventKindOptions = [
		{
			label: "Observation",
			value: "observation",
			icon: mdiMagnify,
			hint: "What was noticed or detected",
		},
		{
			label: "Action",
			value: "action",
			icon: mdiExclamation,
			hint: "Steps taken or changes made",
		},
		{
			label: "Decision",
			value: "decision",
			icon: mdiBook,
			hint: "Choices made and their rationale",
		},
		{
			label: "Context",
			value: "context",
			icon: mdiBrain,
			hint: "Background information or ongoing conditions",
		},
	];

	let descriptionEditor = $state<ReturnType<typeof createMentionEditor>>();
	onMount(() => {
		descriptionEditor = createMentionEditor("", "cursor-text focus:outline-none min-h-20");
		return () => (descriptionEditor?.destroy())
	});
</script>

<div class="flex flex-col gap-2 flex-1">
	<TextField label="Title" value="" />

	<Field label="Event Kind">
		<ToggleGroup bind:value={eventAttributes.eventKind} variant="fill" inset class="w-full">
			{#each eventKindOptions as opt}
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

	<Field label="Key Event" let:id icon={mdiFlag}>
		<Switch {id} bind:value={eventAttributes.isKey} />
	</Field>

	<DateTimePickerField
		label="Time"
		current={dateAnchor}
		onChange={newDate => {
			dateAnchor = newDate;
		}}
		exactTime
	/>

	<Field label="Description" classes={{ root: "grow", container: "h-full", input: "block" }}>
		{#if descriptionEditor}
			<EditorContent editor={descriptionEditor} />
		{/if}
	</Field>
</div>
