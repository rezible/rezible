<script lang="ts">
	import {
	listOncallAnnotationsOptions,
		updateOncallShiftHandoverMutation,
		type ListOncallAnnotationsData,
		type OncallAnnotation,
		type OncallEvent,
		type OncallShiftHandover,
		type UpdateOncallShiftHandoverRequestBody,
	} from "$lib/api";
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import EventRow from "$components/oncall-events/EventRow.svelte";
	import { useShiftViewState } from "../shiftViewState.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {
		handover: OncallShiftHandover;
		onUpdated: () => void;
	};
	const { handover, onUpdated }: Props = $props();

	const viewState = useShiftViewState();

	const annotationsQueryOptions = $derived<ListOncallAnnotationsData["query"]>({
		shiftId: handover.attributes.shiftId,
	});
	const annotationsQuery = createQuery(() => listOncallAnnotationsOptions({ query: annotationsQueryOptions }));
	const annotations = $derived(annotationsQuery.data?.data ?? []);

	const pinnedAnnos = $derived(handover.attributes.pinnedAnnotations ?? []);
	const pinnedEventIds = $derived(new Set(pinnedAnnos.map(p => p.attributes.event.id)));

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
		const ids = new Set(pinnedAnnos.map(a => a.id));

		// toggle in set
		if (!ids.delete(anno.id)) ids.add(anno.id);

		const body: UpdateOncallShiftHandoverRequestBody = {
			attributes: {pinnedAnnotationIds: ids.values().toArray()},
		};
		updateHandoverMut.mutate({ path: { id: handover.id }, body });
	};

	type EventAnnoListItem = {
		event: OncallEvent;
		anno: OncallAnnotation;
		pinned: boolean;
	};
	const listItems = $derived.by(() => {
		const items: EventAnnoListItem[] = [];
		annotations.forEach(anno => {
			const eventId = anno.attributes.event.id;
			const event = viewState.eventIdMap.get(eventId);
			if (!!event) items.push({event, anno, pinned: pinnedEventIds.has(eventId)});
		});
		return items;
	});
</script>

<div class="flex flex-col h-full border border-surface-content/10">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Annotated Shift Events" subheading="" />
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#each listItems as item}
			<EventRow event={item.event} annotations={[item.anno]} pinned={item.pinned} {loadingId} togglePinned={() => togglePinned(item.anno)} />
		{:else}
			<div class="grid place-items-center p-4">
				<span>No Events Annotated</span>
			</div>
		{/each}
	</div>
</div>