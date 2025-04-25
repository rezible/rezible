<script lang="ts">
	import {
	listOncallAnnotationsOptions,
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

	const shiftAnnoEventsQuery = createQuery(() => listOncallAnnotationsOptions({ query: { shiftId } }));
	const annos = $derived(shiftAnnoEventsQuery.data?.data ?? []);

	let pinnedAnnos = $derived(handover.attributes.pinnedAnnotations ?? []);
	const pinnedEventIds = $derived(new SvelteSet(pinnedAnnos.map(p => p.attributes.event.id)));

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
	const togglePinned = (anno: OncallAnnotation) => {
		loadingId = $state.snapshot(anno.attributes.event.id);
		const ids = [...pinnedAnnos.map(a => a.id), anno.id];
		const body: UpdateOncallShiftHandoverRequestBody = {
			attributes: {pinnedAnnotationIds: ids},
		};
		updateHandoverMut.mutate({ path: { id: handover.id }, body });
	};

	const blankAnnotation: OncallAnnotation = {
		id: "foo", 
		attributes: {
			creator: { id: "", attributes: { email: "", name: "" } },
			event: {
				id: "",
				attributes: {
					annotations: [],
					description: "",
					kind: "",
					timestamp: "",
					title: ""
				}
			},
			minutesOccupied: 0,
			notes: "bleh",
			roster: {
				id: "",
				attributes: {
					handoverTemplateId: "",
					name: "",
					schedules: [],
					slug: ""
				}
			},
			tags: []
		}
	}
</script>

<div class="flex flex-col h-full border border-surface-content/20 rounded-lg">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Annotated Shift Events" subheading="" />
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#each pinnedAnnos as annotation}
			{@const event = annotation.attributes.event}
			<EventRowItem {event} {annotation} pinned {loadingId} togglePinned={() => togglePinned(annotation)} />
		{/each}

		{#each annos as annotation}
			{@const event = annotation.attributes.event}
			{#if !pinnedEventIds.has(event.id)}
				<EventRowItem {event} {annotation} {loadingId} togglePinned={() => togglePinned(annotation)} />
			{/if}
		{/each}
	</div>
</div>