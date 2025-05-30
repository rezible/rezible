<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import { Button, Checkbox, Field, Icon, TextField, ToggleGroup, ToggleOption } from "svelte-ux";
	import { attributesState } from "./attributes.svelte";
	import { mdiCalendar, mdiClose, mdiPlus } from "@mdi/js";
	import { settings } from "$lib/settings.svelte";
	import { PeriodType } from "@layerstack/utils";

	type Props = {
		event: OncallEvent;
		current?: OncallAnnotation;
	};
	const { event, current }: Props = $props();

	attributesState.setup(event, current);

	const eventTimeFmt = $derived(settings.format(event.attributes.timestamp, PeriodType.DayTime));

	const kind = $derived(event.attributes.kind);

	const onTagsInputKeypress = (e: KeyboardEvent) => {
		if (e.key === "Enter") {
			attributesState.addDraftTag();
		}
	};

	const addTagButtonDisabled = $derived(attributesState.tags.has(attributesState.draftTag));
</script>

<div class="flex flex-col gap-2">
	<div class="flex flex-col">
		<span class="text-lg">{event.attributes.title}</span>
		<span class="flex items-center gap-1 text-surface-content/70">
			<Icon data={mdiCalendar} size={20} />
			{eventTimeFmt}
		</span>
		<span class="text-surface-content/50">{event.attributes.description}</span>
	</div>

	{#if kind === "alert"}
		<Field label="Did this alert indicate a real issue?">
			<ToggleGroup inset variant="fill" size="lg" bind:value={attributesState.alertAccuracy}>
				<ToggleOption value="yes">Yes</ToggleOption>
				<ToggleOption value="no">No</ToggleOption>
				<ToggleOption value="unknown">Unknown</ToggleOption>
			</ToggleGroup>
		</Field>

		<Field label="Did this alert require action to be taken?">
			<ToggleGroup inset variant="fill" size="lg" bind:value={attributesState.alertRequiredAction}>
				<ToggleOption value={true}>Yes</ToggleOption>
				<ToggleOption value={false}>No</ToggleOption>
			</ToggleGroup>
		</Field>

		<Field label="Was documentation available?" classes={{ input: "gap-3" }}>
			<ToggleGroup
				inset
				variant="fill"
				size="lg"
				value={attributesState.alertDocs}
				on:change={(e) => {
					attributesState.alertDocs = !!e.detail.value;
				}}
			>
				<ToggleOption value={true}>Yes</ToggleOption>
				<ToggleOption value={false}>No</ToggleOption>
			</ToggleGroup>

			{#if attributesState.alertDocs}
				<Checkbox bind:value={attributesState.alertDocsNeedUpdate}>Needs Update?</Checkbox>
			{/if}
		</Field>
	{/if}

	<Field label="Tags">
		<div class="flex gap-2" class:mr-2={attributesState.tags.size > 0}>
			{#each attributesState.tags as tag}
				<div class="flex items-center gap-1 border rounded-lg p-1 px-2">
					<span class="leading-4">{tag}</span>
					<Button
						size="sm"
						icon={mdiClose}
						iconOnly
						on:click={() => {
							attributesState.removeTag(tag);
						}}
					/>
				</div>
			{/each}
		</div>
		<TextField
			placeholder="Add Tag"
			on:keydown={onTagsInputKeypress}
			bind:value={attributesState.draftTag}
		>
			<span slot="append">
				<Button
					icon={mdiPlus}
					iconOnly
					disabled={addTagButtonDisabled}
					class="text-surface-content/50 p-2"
					on:click={() => {
						attributesState.addDraftTag();
					}}
				/>
			</span>
		</TextField>
	</Field>

	<TextField label="Notes" multiline bind:value={attributesState.notes} classes={{ input: "h-28" }} />
</div>
