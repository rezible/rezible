<script lang="ts">
	import {
		listOncallEventsOptions,
		updateOncallShiftHandoverMutation,
		type OncallAnnotation,
		type OncallShiftHandover,
		type UpdateOncallShiftHandoverRequestBody,
	} from "$lib/api";
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import { SvelteSet } from "svelte/reactivity";
	import { Header } from "svelte-ux";
	import EventRowItem from "$components/oncall-events/EventRowItem.svelte";
	import { shiftViewStateCtx } from "../context.svelte";

	type Props = {
		handover: OncallShiftHandover;
		onUpdated: () => void;
	};
	const { handover, onUpdated }: Props = $props();

	const viewState = shiftViewStateCtx.get();
	const shiftId = $derived(viewState.shiftId);

	const annoEventsQuery = createQuery(() => listOncallEventsOptions({ query: { shiftId, annotated: true } }));
	const events = $derived(annoEventsQuery.data?.data ?? []);

	let pinnedAnnos = $derived(handover.attributes.pinnedEvents ?? []);
	const pinnedEventIds = $derived(new SvelteSet(pinnedAnnos.map(p => p.event.id)));

	let loadingId = $state<string>();
	const updateHandoverMut = createMutation(() => ({
		...updateOncallShiftHandoverMutation(),
		onSuccess: () => {
			// pinnedAnnos = data.data.attributes.pinnedEvents;
			onUpdated();
		},
		onSettled: () => {
			loadingId = undefined;
		}
	}));
	const togglePinned = (eventId: string, annoId: string) => {
		loadingId = $state.snapshot(eventId);
		const ids = [...pinnedAnnos.map(a => a.annotation.id), annoId];
		const body: UpdateOncallShiftHandoverRequestBody = {
			attributes: {pinnedAnnotationIds: ids},
		};
		updateHandoverMut.mutate({ path: { id: handover.id }, body });
	};

	const blankAnnotation: OncallAnnotation = {id: "foo", attributes: {
		creator: {id: "", attributes: {email: "", name: ""}},
		eventId: "",
		minutesOccupied: 0,
		notes: "bleh",
		rosterId: ""
	}}
</script>

<div class="flex flex-col h-full border border-surface-content/10 rounded">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Annotated Shift Events" subheading="" />
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#each pinnedAnnos as {event, annotation}}
			<EventRowItem {event} {annotation} pinned {loadingId} togglePinned={() => togglePinned(event.id, annotation.id)} />
		{/each}

		{#each events as event}
			{@const annotation = event.attributes.annotations?.at(0) ?? blankAnnotation}
			{#if !pinnedEventIds.has(event.id)}
				<EventRowItem {event} {annotation} {loadingId} togglePinned={() => togglePinned(event.id, annotation.id)} />
			{/if}
		{/each}
	</div>
</div>