<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import { Checkbox, Field, TextField, ToggleGroup, ToggleOption } from "svelte-ux";
	import Button from "$components/button/Button.svelte";
	import { mdiClose, mdiPlus } from "@mdi/js";
	import { useAnnotationDialogState } from "./dialogState.svelte";

	type Props = {
		event: OncallEvent;
		current?: OncallAnnotation;
	};
	const { event, current }: Props = $props();

	const dialog = useAnnotationDialogState();

	const kind = $derived(event.attributes.kind);

	const onTagsInputKeypress = (e: KeyboardEvent) => {
		if (e.key === "Enter") {
			dialog.attributes.addDraftTag();
		}
	};

	const addTagButtonDisabled = $derived(dialog.attributes.tags.has(dialog.attributes.draftTag));
</script>

<div class="flex flex-col gap-2">
	{#if kind === "alert"}
		<Field label="Did this alert indicate a real issue?">
			<ToggleGroup inset variant="fill" size="lg" bind:value={dialog.attributes.alertAccuracy}>
				<ToggleOption value="yes">Yes</ToggleOption>
				<ToggleOption value="no">No</ToggleOption>
				<ToggleOption value="unknown">Unknown</ToggleOption>
			</ToggleGroup>
		</Field>

		<Field label="Did this alert require action to be taken?">
			<ToggleGroup inset variant="fill" size="lg" bind:value={dialog.attributes.alertRequiredAction}>
				<ToggleOption value={true}>Yes</ToggleOption>
				<ToggleOption value={false}>No</ToggleOption>
			</ToggleGroup>
		</Field>

		<Field label="Was documentation available?" classes={{ input: "gap-3" }}>
			<ToggleGroup
				inset
				variant="fill"
				size="lg"
				value={dialog.attributes.alertDocs}
				on:change={(e) => {
					dialog.attributes.alertDocs = !!e.detail.value;
				}}
			>
				<ToggleOption value={true}>Yes</ToggleOption>
				<ToggleOption value={false}>No</ToggleOption>
			</ToggleGroup>

			{#if dialog.attributes.alertDocs}
				<Checkbox bind:value={dialog.attributes.alertDocsNeedUpdate}>Needs Update?</Checkbox>
			{/if}
		</Field>
	{/if}

	<Field label="Tags">
		<div class="flex gap-2" class:mr-2={dialog.attributes.tags.size > 0}>
			{#each dialog.attributes.tags as tag}
				<div class="flex items-center gap-1 border rounded-lg p-1 px-2">
					<span class="leading-4">{tag}</span>
					<Button
						size="sm"
						icon={mdiClose}
						iconOnly
						on:click={() => {
							dialog.attributes.removeTag(tag);
						}}
					/>
				</div>
			{/each}
		</div>
		<TextField
			placeholder="Add Tag"
			on:keydown={onTagsInputKeypress}
			bind:value={dialog.attributes.draftTag}
		>
			<span slot="append">
				<Button
					icon={mdiPlus}
					iconOnly
					disabled={addTagButtonDisabled}
					class="text-surface-content/50 p-2"
					on:click={() => {
						dialog.attributes.addDraftTag();
					}}
				/>
			</span>
		</TextField>
	</Field>

	<TextField label="Notes" multiline bind:value={dialog.attributes.notes} classes={{ input: "h-28" }} />
</div>
