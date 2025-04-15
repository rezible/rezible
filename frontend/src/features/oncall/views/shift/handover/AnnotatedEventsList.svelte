<script lang="ts">
	import {
		listOncallEventAnnotationsOptions,
		updateOncallAnnotationMutation,
		type OncallEvent,
		type OncallEventAnnotation,
		type OncallShift,
	} from "$lib/api";
	import { mdiPin, mdiPinOutline, mdiDotsVertical, mdiCircleMedium } from "@mdi/js";
	import { Icon, Button, Header, Toggle, Menu, MenuItem } from "svelte-ux";
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { SvelteSet } from "svelte/reactivity";
	import { settings } from "$lib/settings.svelte";
	import { PeriodType } from "@layerstack/utils";

	type Props = {
		shift: OncallShift;
		editable: boolean;
		pinnedAnnotations: OncallEventAnnotation[];
	};
	const { shift, editable, pinnedAnnotations }: Props = $props();

	const annoEventsQuery = createQuery(() => listOncallEventAnnotationsOptions({ query: { shiftId: shift.id } }));

	const annoEvents = $derived(annoEventsQuery.data?.data ?? []);

	const pinnedEventIds = $derived(new SvelteSet(pinnedAnnotations.map(p => p.event.id)));
	const unpinnedAnnos = $derived(annoEvents.filter(a => (!pinnedEventIds.has(a.event.id))));

	const updateAnnotationMut = createMutation(() => ({
		...updateOncallAnnotationMutation(),
	}));
	const togglePinned = (event: OncallEvent) => {
		// const body: UpdateOncallAnnotationRequestBody = {
		// 	attributes: {
		// 		pinned: !ann.attributes.pinned,
		// 	},
		// };
		// updateAnnotationMut.mutate({ path: { id: ann.id }, body });
	};
</script>

<div class="h-10 flex w-full gap-4 items-center px-2">
	<Header title="Annotated Shift Events" subheading="All events with annotations for this shift" classes={{ root: "w-full", title: "text-xl", container: "flex-1" }}>
		<div slot="actions" class:hidden={!editable}>
			
		</div>
	</Header>
</div>

<div class="flex-1 min-h-0 flex flex-col gap-4 overflow-y-auto bg-surface-200 p-3">
	{#if annoEventsQuery.isLoading}
		<span>loading</span>
	{:else if annoEvents.length > 0}
		{#each pinnedAnnotations as ev, i}
			{@render eventListItem(ev, true)}
		{/each}

		{#each unpinnedAnnos as ev, i}
			{@render eventListItem(ev, false)}
		{/each}
	{:else}
		<span>No Annotations</span>
	{/if}
</div>

{#snippet eventListItem(eventAnno: OncallEventAnnotation, pinned: boolean)}
	{@const event = eventAnno.event}
	{@const occurredAt = event.attributes.timestamp ?? ""}
	<div class="grid grid-cols-[100px_auto_minmax(0,1fr)] place-items-center border p-2">
		<div class="justify-self-start">
			<span class="flex items-center">
				{settings.format(occurredAt, PeriodType.DayTime)}
			</span>
		</div>

		<div class="items-center static z-10">
			<Icon
				data={mdiCircleMedium}
				classes={{ root: "bg-accent-900 rounded-full p-2 w-auto h-10" }}
			/>
		</div>

		<div class="w-full justify-self-start grid grid-cols-[auto_40px] items-center px-2">
			<div class="leading-none">{event.attributes.title || "todo: event title"}</div>
			<div class="place-self-end flex flex-row gap-2" class:hidden={!editable}>
				<Button
					disabled={updateAnnotationMut.isPending}
					icon={pinned ? mdiPin : mdiPinOutline}
					color={pinned ? "accent" : "default"}
					iconOnly
					on:click={() => togglePinned(event)}
				/>

				<Toggle let:on={open} let:toggle let:toggleOff>
					<Button on:click={toggle} icon={mdiDotsVertical} iconOnly>
						<Menu {open} on:close={toggleOff}>
							<MenuItem>Edit</MenuItem>
							<MenuItem>Delete</MenuItem>
						</Menu>
					</Button>
				</Toggle>
			</div>
		</div>

		{#each event.attributes.annotations as anno}
			<div class="row-start-3 col-start-3 overflow-y-auto max-h-20 overflow-y-auto border rounded p-2 w-full">
				{anno.attributes.notes}
			</div>
		{/each}
	</div>
{/snippet}
