<script lang="ts">
	import {
		listOncallAnnotationsOptions,
		listOncallEventsOptions,
		updateOncallAnnotationMutation,
		type OncallAnnotation,
		type OncallEvent,
		type OncallShift,
	} from "$lib/api";
	import { mdiPlus, mdiPin, mdiPinOutline, mdiDotsVertical, mdiCircleMedium } from "@mdi/js";
	import { Icon, Button, Header, Toggle, Menu, MenuItem } from "svelte-ux";
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { SvelteSet } from "svelte/reactivity";
	import { settings } from "$lib/settings.svelte";
	import ShiftAnnotationEditorDialog from "./ShiftAnnotationEditorDialog.svelte";
	import { PeriodType } from "@layerstack/utils";

	type Props = {
		shift: OncallShift;
		editable: boolean;
		pinnedEvents: OncallEvent[];
	};
	const { shift, editable, pinnedEvents }: Props = $props();

	const queryClient = useQueryClient();
	const eventsQueryOpts = $derived(listOncallEventsOptions({ query: { shiftId: shift.id } }));
	const eventsQuery = createQuery(() => eventsQueryOpts);
	const invalidateEventsQuery = () => queryClient.invalidateQueries(eventsQueryOpts);

	const events = $derived(eventsQuery.data?.data ?? []);

	const pinnedEventIds = $derived(new SvelteSet(pinnedEvents.map(e => e.id)));
	const unpinnedEvents = $derived(events.filter(a => (!pinnedEventIds.has(a.id))));

	let showEditorDialog = $state(false);

	const onAnnotationCreated = () => {
		invalidateEventsQuery();
		showEditorDialog = false;
	};

	const updateAnnotationMut = createMutation(() => ({
		...updateOncallAnnotationMutation(),
		onSuccess: invalidateEventsQuery,
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
	<Header title="Shift Events" classes={{ root: "w-full", title: "text-xl", container: "flex-1" }}>
		<div slot="actions" class:hidden={!editable}>
			
		</div>
	</Header>
</div>

<div class="flex-1 min-h-0 flex flex-col gap-4 overflow-y-auto bg-surface-200 p-3">
	{#if eventsQuery.isLoading}
		<span>loading</span>
	{:else if events.length > 0}
		{#if editable}
			<Header title="Pinned" subheading="Included in the handover notes" />
		{/if}

		{#each pinnedEvents as ev, i}
			{@render eventListItem(ev, true)}
		{:else}
			<span>Nothing Pinned</span>
		{/each}

		{#if unpinnedEvents.length > 0}
			<div class="w-full border-b"></div>

			{#each unpinnedEvents as ev, i}
				{@render eventListItem(ev, false)}
			{/each}
		{/if}
	{:else}
		<span>No Annotations</span>
	{/if}
</div>

{#snippet eventListItem(event: OncallEvent, pinned: boolean)}
	{@const anno = event.attributes.annotation}
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
					loading={updateAnnotationMut.isPending &&
						updateAnnotationMut.variables?.path.id === anno.id}
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

		<div
			class="row-start-3 col-start-3 overflow-y-auto max-h-20 overflow-y-auto border rounded p-2 w-full"
		>
			{anno.attributes.notes}
		</div>
	</div>
{/snippet}

<!--div class="pt-2 min-h-0 flex-1 overflow-y-auto bg-surface-200 flex">
	<div class="h-full max-h-full min-h-0 w-40">
		<span>timeline</span>
	</div>

	<div class="flex-1 grid auto-rows-min grid-cols-[100px_auto_minmax(0,1fr)] overflow-y-auto">
		<div class="grid grid-cols-subgrid col-span-3 grid-rows-[40px_20px] row-span-2 gap-x-2 p-2">
			<div class="relative col-start-2 col-span-1 place-items-center grid">
				<Icon class="absolute -top-2 text-accent-900" data={mdiCircleMedium} size={32} />
			</div>

			<div class="col-start-2 row-start-2 grid place-items-center -mt-8">
				<hr class="w-1 z-10 h-20 justify-self-center bg-accent-900 border-accent-900" />
			</div>
		</div>

		{#snippet timelineEvent(event: ShiftTimelineEvent, isFlagged: boolean)}
			<div class="grid grid-cols-subgrid col-span-3 place-items-center">
				<div class="justify-self-start">
					<span class="flex items-center">
						{event.occurredAt.toLocaleTimeString()}
					</span>
				</div>

				<div class="items-center static z-10">
					<Icon data={eventKindIcons[event.kind]} classes={{root: "bg-accent-900 rounded-full p-2 w-auto h-10"}} />
				</div>

				<div class="w-full justify-self-start grid grid-cols-[auto_40px] items-center">
					<div class="leading-none">{event.title}</div>
					<div class="place-self-end">
						<Button 
							icon={isFlagged ? mdiFlag : mdiFlagOutline}
							color={isFlagged ? "accent" : "default"}
							iconOnly 
							on:click={() => {handoverState.flagEvent(event)}} />
					</div>
				</div>
			</div>

			<div class="col-start-2 relative row-start-2 min-h-10 h-[var(--distance-height)] place-items-center grid">
				<hr class="absolute w-1 top-0 h-[calc(100%+24px)] justify-self-center bg-accent-900 border-accent-900" />
			</div>

			{#if event.description}
				<div class="col-start-3 row-start-2 overflow-y-auto max-h-20 overflow-y-auto border rounded p-2">
					<div class=""></div>
				</div>
			{/if}
		{/snippet}

		{#each timeline as node, i}
			{@const style = `--distance-height: ${node.height}px`}
			{@const event = node.event}
			{#if event}
				{@const isFlagged = handoverState.flaggedEvents[event.eventId]}
				<div class="grid grid-cols-subgrid col-span-3 grid-rows-[40px_auto] row-span-2 gap-x-2 p-2 {isFlagged ? "bg-accent-800/20" : ""}" {style}>
					{@render timelineEvent(event, isFlagged)}
				</div>
			{/if}
		{/each}

		<div class="grid grid-cols-subgrid col-span-3 grid-rows-[40px_auto] row-span-2 gap-x-2 p-2">
			<div class="justify-self-start">
				<span class="flex items-center">
					
				</span>
			</div>

			<div class="col-start-2 col-span-1 place-items-start grid -mt-3">
				<Icon class="text-accent-900 w-full" data={mdiCircleMedium} size={36} />
			</div>

			<div class="row-start-2 col-start-2 place-items-center grid -mt-2">
				Shift Started
			</div>
		</div>
	</div>
</div-->
