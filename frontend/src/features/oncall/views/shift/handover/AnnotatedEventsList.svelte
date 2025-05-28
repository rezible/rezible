<script lang="ts">
	import {
		listOncallAnnotationsOptions,
		updateOncallShiftHandoverMutation,
		type OncallAnnotation,
		type OncallShiftHandover,
		type UpdateOncallShiftHandoverRequestBody,
	} from "$lib/api";
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import { Header } from "svelte-ux";
	import EventRowItem from "$components/oncall-events/EventRowItem.svelte";
	import { useShiftViewState } from "../shiftViewState.svelte";

	type Props = {
		handover: OncallShiftHandover;
		onUpdated: () => void;
	};
	const { handover, onUpdated }: Props = $props();

	const viewState = useShiftViewState();
	const shiftId = $derived(viewState.shiftId);

	const shiftAnnoEventsQuery = createQuery(() => listOncallAnnotationsOptions({ query: { shiftId } }));
	const annos = $derived(shiftAnnoEventsQuery.data?.data ?? []);

	const pinnedAnnos = $derived(handover.attributes.pinnedAnnotations ?? []);
	const pinnedEventIds = $derived(new Set(pinnedAnnos.map(p => p.attributes.event.id)));
	const unpinnedAnnos = $derived(annos.filter(a => !pinnedEventIds.has(a.attributes.event.id)));

	let loadingId = $state<string>();
	const updateHandoverMut = createMutation(() => ({
		...updateOncallShiftHandoverMutation(),
		onSuccess: () => {
			onUpdated();
		},
		onSettled: () => {
			loadingId = undefined;
		}
	}));
	const togglePinned = (anno: OncallAnnotation) => {
		loadingId = $state.snapshot(anno.attributes.event.id);
		const ids = [...pinnedAnnos.map(a => a.id), anno.id];
		const body: UpdateOncallShiftHandoverRequestBody = {
			attributes: {pinnedAnnotationIds: ids},
		};
		updateHandoverMut.mutate({ path: { id: handover.id }, body });
	};
</script>

<div class="flex flex-col h-full border border-surface-content/10">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Annotated Shift Events" subheading="" />
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#each pinnedAnnos as annotation}
			{@const event = annotation.attributes.event}
			<EventRowItem {event} {annotation} pinned {loadingId} togglePinned={() => togglePinned(annotation)} />
		{/each}

		{#each unpinnedAnnos as annotation}
			{@const event = annotation.attributes.event}
			<EventRowItem {event} {annotation} {loadingId} togglePinned={() => togglePinned(annotation)} />
		{/each}

		{#if annos.length === 0}
			<div class="grid place-items-center p-4">
				<span>No Events Annotated</span>
			</div>
		{/if}
	</div>
</div>