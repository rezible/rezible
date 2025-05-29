<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import { mdiPlus } from "@mdi/js";
	import { Button, Icon } from "svelte-ux";
	import EventTime from "./EventTime.svelte";
	import EventRowIcon from "./EventRowIcon.svelte";

	type Props = {
		event: OncallEvent;
		allowAnnotationRosters: string[];
		onOpenAnnotateDialog: (anno?: OncallAnnotation) => void;
	}
	let { event, allowAnnotationRosters, onOpenAnnotateDialog }: Props = $props();

	const attrs = $derived(event.attributes);

	const annoRosters = $derived(new Set(event.attributes.annotations.map(a => a.attributes.roster.id)));
	const needsAnnotation = $derived(allowAnnotationRosters.some(id => !annoRosters.has(id)));
</script>

<div class="group grid grid-cols-subgrid col-span-full hover:bg-surface-100/50 h-16 p-2">
	<EventRowIcon kind={attrs.kind} />

	<div class="grid items-center">
		<EventTime timestamp={event.attributes.timestamp} />
	</div>
	
	<div class="flex flex-col h-full">
		<div class="font-medium flex items-center">
			{attrs.title || `${attrs.kind.charAt(0).toUpperCase() + attrs.kind.slice(1)} ${event.id.substring(0, 8)}`}
		</div>
		<div class="text-sm">
			description
		</div>
	</div>

	<div class="flex items-center justify-end">
		{#each event.attributes.annotations as anno}
			<div class="border p-1">
				<span>{anno.attributes.notes}</span>
				{#if allowAnnotationRosters.includes(anno.attributes.roster.id)}
					<Button variant="fill-light" on:click={() => (onOpenAnnotateDialog(anno))}>
						<span class="flex items-center gap-2">
							Edit
						</span>
					</Button>
				{/if}
			</div>
		{/each}

		{#if needsAnnotation}
			<div class="hidden group-hover:inline">
				<Button variant="fill-light" color={"accent"} on:click={() => (onOpenAnnotateDialog())}>
					<span class="flex items-center gap-2">
						Annotate
						<Icon data={mdiPlus} />
					</span>
				</Button>
			</div>
		{/if}
	</div>
</div>
