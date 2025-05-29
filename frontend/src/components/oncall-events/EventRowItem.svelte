<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import { mdiPin, mdiPinOutline } from "@mdi/js";
	import { Button } from "svelte-ux";
	import EventTime from "./EventTime.svelte";
	import EventRowIcon from "./EventRowIcon.svelte";

	type Props = {
		event: OncallEvent;
		annotations?: OncallAnnotation[];
		annotatableRosterIds?: string[];
		editAnnotation?: (anno?: OncallAnnotation) => void;
		pinned?: boolean;
		togglePinned?: () => void;
		loadingId?: string;
	}
	const { event, annotations = [], annotatableRosterIds = [], editAnnotation, pinned, togglePinned, loadingId }: Props = $props();

	const attrs = $derived(event.attributes);

	const rosterIdsWithAnnotations = $derived(new Set(annotations.map(a => a.attributes.roster.id)));
	const showAnnotationButton = $derived(annotatableRosterIds.some(id => !rosterIdsWithAnnotations.has(id)));
	const loading = $derived(!!loadingId && loadingId === event.id);
	const disabled = $derived(!!loadingId && loadingId !== event.id);
</script>

<div class="group grid grid-cols-[auto_auto_auto_minmax(0,1fr)] gap-4 place-items-center border p-3 bg-neutral-900/40 border-neutral-content/10 shadow-sm hover:shadow-md transition-shadow">
	<div class="justify-self-end flex flex-col items-start">
		<EventTime timestamp={attrs.timestamp} />
	</div>

	<EventRowIcon kind={attrs.kind} />

	<div class="flex flex-col h-full">
		<div class="font-medium flex items-center">
			{attrs.title || `${attrs.kind.charAt(0).toUpperCase() + attrs.kind.slice(1)} ${event.id.substring(0, 8)}`}
		</div>
		<div class="text-sm">
			description
		</div>
	</div>

	<div class="flex items-center gap-4 w-full">
		{#each annotations as anno}
			<div class="overflow-y-auto w-full h-full border rounded p-2 bg-neutral-700/70 text-sm flex items-center">
				<div class="text-neutral-content">{anno.attributes.notes}</div>
			</div>
		{:else} 
			{#if editAnnotation && showAnnotationButton}
				<div class="hidden group-hover:inline w-full h-full">
					<Button classes={{root: "w-full h-full"}} {loading} {disabled} on:click={() => editAnnotation()}>Add Annotation</Button>
				</div>
			{/if}
		{/each}

		{#if !!togglePinned}
			<div class="pl-4">
				<Button iconOnly icon={pinned ? mdiPin : mdiPinOutline} {loading} {disabled} on:click={togglePinned} />
			</div>
		{/if}
	</div>
</div>