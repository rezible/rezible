<script lang="ts">
	import TiptapEditor from "$components/tiptap-editor/TiptapEditor.svelte";
	import {
		Field,
		ToggleGroup,
		ToggleOption,
		Tooltip,
		Icon,
		TextField,
		Switch,
	} from "svelte-ux";
	import { mdiMagnify, mdiExclamation, mdiBook, mdiBrain, mdiFlag } from "@mdi/js";
	import { onMount } from "svelte";
	import DateTimePickerField from "$components/date-time-field/DateTimePickerField.svelte";
	import { useEventDialog } from "../dialogState.svelte";
	
	const eventDialog = useEventDialog();

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

	onMount(eventDialog.eventAttributes.mountDescriptionEditor);
</script>

<div class="flex flex-col gap-2 flex-1">
	<TextField label="Title" bind:value={eventDialog.eventAttributes.title} />

	<Field label="Event Kind">
		<ToggleGroup bind:value={eventDialog.eventAttributes.kind} variant="fill" inset class="w-full">
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
		<Switch {id} bind:value={eventDialog.eventAttributes.isKey} />
	</Field>

	<DateTimePickerField
		label="Time"
		current={eventDialog.eventAttributes.timestamp}
		onChange={ts => (eventDialog.eventAttributes.timestamp = ts)}
		exactTime
	/>

	<Field label="Description" classes={{ root: "grow", container: "h-full", input: "block" }}>
		{#if eventDialog.eventAttributes.descriptionEditor}
			<TiptapEditor bind:editor={eventDialog.eventAttributes.descriptionEditor} />
		{/if}
	</Field>
</div>
