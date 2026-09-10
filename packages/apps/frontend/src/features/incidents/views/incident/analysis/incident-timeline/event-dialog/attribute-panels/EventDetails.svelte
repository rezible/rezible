<script lang="ts">
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import RiErrorWarningLine from "remixicon-svelte/icons/error-warning-line";
	import RiBookOpenLine from "remixicon-svelte/icons/book-open-line";
	import RiBrainLine from "remixicon-svelte/icons/brain-line";
	import RiFlagLine from "remixicon-svelte/icons/flag-line";
	import { onMount } from "svelte";
	import DateTimePickerField from "$src/components/forms/date-time-field/DateTimePickerField.svelte";
	import { useEventDialogAttributes } from "./attributes.svelte";

	const attributes = useEventDialogAttributes();

	const eventKindOptions = [
		{
			label: "Observation",
			value: "observation",
			icon: RiSearchLine,
			hint: "What was noticed or detected",
		},
		{
			label: "Action",
			value: "action",
			icon: RiErrorWarningLine,
			hint: "Steps taken or changes made",
		},
		{
			label: "Decision",
			value: "decision",
			icon: RiBookOpenLine,
			hint: "Choices made and their rationale",
		},
		{
			label: "Context",
			value: "context",
			icon: RiBrainLine,
			hint: "Background information or ongoing conditions",
		},
	];

	onMount(attributes.mountDescriptionEditor);
</script>

<div class="flex flex-col gap-2 flex-1">
	<span>title field</span>
	<!-- <TextField label="Title" bind:value={eventAttributes.title} /> -->

	<DateTimePickerField
		label="Time"
		current={attributes.timestamp}
		onChange={(ts) => (attributes.timestamp = ts)}
		exactTime
	/>

	<span>kind field</span>
	<!-- <Field label="Event Kind">
		<ToggleGroup bind:value={eventAttributes.kind} variant="fill" inset class="w-full">
			{#each eventKindOptions as opt}
				<ToggleOption value={opt.value}>
					<Tooltip title={opt.hint}>
						<span class="flex items-center justify-center gap-2 px-2">
							<svelte:component this={opt.icon} aria-hidden="true" />
							{opt.label}
						</span>
					</Tooltip>
				</ToggleOption>
			{/each}
		</ToggleGroup>
	</Field> -->

	<span>key event toggle</span>
	<!-- <Field label="Key Event" let:id icon={RiFlagLine}>
		<Switch {id} bind:value={eventAttributes.isKey} />
	</Field> -->

	<span>desc field</span>
	<!-- <Field label="Description" classes={{ root: "grow", container: "h-full", input: "block" }}>
		{#if eventAttributes.descriptionEditor}
			<TiptapEditor bind:editor={eventAttributes.descriptionEditor} />
		{/if}
	</Field> -->
</div>
