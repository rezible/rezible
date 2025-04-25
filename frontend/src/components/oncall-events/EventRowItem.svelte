<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import { mdiPin, mdiPinOutline } from "@mdi/js";
	import { Button } from "svelte-ux";
	import EventTime from "./EventTime.svelte";
	import EventRowIcon from "./EventRowIcon.svelte";

	type Props = {
		event: OncallEvent;
		annotation?: OncallAnnotation;
		annotationRosterIds?: string[];
		editAnnotation?: (anno?: OncallAnnotation) => void;
		pinned?: boolean;
		togglePinned?: () => void;
		loadingId?: string;
	}
	const { event, annotation, annotationRosterIds, editAnnotation, pinned, togglePinned, loadingId }: Props = $props();

	const attrs = $derived(event.attributes);

	const annotations = $derived(annotationRosterIds ? event.attributes.annotations.filter(a => annotationRosterIds?.includes(a.id)) : []);
	const rosterIdsWithAnnotations = $derived(new Set(event.attributes.annotations.map(a => a.attributes.roster.id)));
	const showAnnotationButton = $derived(annotationRosterIds?.some(id => !rosterIdsWithAnnotations.has(id)));
	const loading = $derived(!!loadingId && loadingId === event.id);
	const disabled = $derived(!!loadingId && loadingId !== event.id);
</script>

<div class="group grid grid-cols-[auto_minmax(0,1fr)_auto_auto] place-items-center border p-3 bg-neutral-900/40 border-neutral-content/10 shadow-sm hover:shadow-md transition-shadow">
	<EventRowIcon kind={attrs.kind} />

	<div class="w-full justify-self-start grid grid-cols-[auto_1fr] items-start gap-4 px-4">
		<div class="flex flex-col h-full">
			<div class="font-medium flex items-center">
				{attrs.title || `${attrs.kind.charAt(0).toUpperCase() + attrs.kind.slice(1)} ${event.id.substring(0, 8)}`}
			</div>
			<div class="text-sm">
				description
			</div>
		</div>

		{#if annotation}
			<div class="overflow-y-auto w-full h-full border rounded p-2 bg-neutral-700/70 text-sm flex items-center">
				<div class="text-neutral-content">{annotation.attributes.notes}</div>
			</div>
		{:else}
			{#each annotations as anno}
				<span>roster annotation</span>
			{/each}
		{/if}
		
		{#if editAnnotation && showAnnotationButton}
			<div class="hidden group-hover:inline mx-4 h-full">
				<Button classes={{root: "w-full h-full"}} {loading} {disabled} on:click={() => editAnnotation()}>Add Annotation</Button>
			</div>
		{/if}
	</div>

	<div class="justify-self-end flex flex-col items-start">
		<EventTime timestamp={attrs.timestamp} />
	</div>

	{#if !!togglePinned}
		<div class="pl-4">
			<Button iconOnly icon={pinned ? mdiPin : mdiPinOutline} {loading} {disabled} on:click={togglePinned} />
		</div>
	{/if}
</div>