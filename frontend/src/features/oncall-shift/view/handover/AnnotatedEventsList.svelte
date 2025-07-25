<script lang="ts">
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import {
	listOncallAnnotationsOptions,
		updateOncallShiftHandoverMutation,
		type ListOncallAnnotationsData,
		type OncallAnnotation,
		type OncallEvent,
		type OncallShiftHandover,
		type UpdateOncallShiftHandoverRequestBody,
	} from "$lib/api";
	import Header from "$components/header/Header.svelte";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventRow from "$components/oncall-events/EventRow.svelte";

	type Props = {
		handover: OncallShiftHandover;
		onUpdated: () => void;
	};
	const { handover, onUpdated }: Props = $props();

	const annotationsQueryOptions = $derived<ListOncallAnnotationsData["query"]>({
		shiftId: handover.attributes.shiftId,
		withEvents: true,
	});
	const annotationsQuery = createQuery(() => listOncallAnnotationsOptions({ query: annotationsQueryOptions }));
	const annotations = $derived(annotationsQuery.data?.data ?? []);

	const pinnedAnnos = $derived(handover.attributes.pinnedAnnotations ?? []);
	const pinnedEventIds = $derived(new Set(pinnedAnnos.map(a => a.attributes.event.id)));

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
</script>

<EventAnnotationDialog />

<div class="flex flex-col h-full border border-surface-content/10">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Annotated Shift Events" subheading="" />
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#each annotations as anno}
			<EventRow 
				event={anno.attributes.event as OncallEvent} 
				annotations={[anno]} 
				pinned={pinnedEventIds.has(anno.attributes.event.id)}
				{loadingId} togglePinned={() => togglePinned(anno)} />
		{:else}
			<div class="grid place-items-center p-4">
				<span>No Events Annotated</span>
			</div>
		{/each}
	</div>
</div>