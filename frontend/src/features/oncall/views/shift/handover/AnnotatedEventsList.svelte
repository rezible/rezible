<script lang="ts">
	import {
		listOncallAnnotationsOptions,
		updateOncallShiftHandoverMutation,
		type OncallAnnotation,
		type OncallEvent,
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
		const ids = [...pinnedAnnos.map(a => a.id), anno.id];
		const body: UpdateOncallShiftHandoverRequestBody = {
			attributes: {pinnedAnnotationIds: ids},
		};
		updateHandoverMut.mutate({ path: { id: handover.id }, body });
	};

	type EventAnnoListItem = {
		event: OncallEvent;
		annotation: OncallAnnotation;
		pinned: boolean;
	};
	// TODO: do this in viewState?
	const listItems = $derived.by(() => {
		const items: EventAnnoListItem[] = [];
		viewState.eventAnnotationsMap.forEach((annotation, eventId) => {
			const event = viewState.eventIdMap.get(eventId);
			const pinned = pinnedEventIds.has(eventId);
			if (!!event) items.push({event, annotation, pinned});
		});
		return items;
	})
</script>

<div class="flex flex-col h-full border border-surface-content/10">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Annotated Shift Events" subheading="" />
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#each listItems as item}
			<EventRowItem {...item} {loadingId} togglePinned={() => togglePinned(item.annotation)} />
		{:else}
			<div class="grid place-items-center p-4">
				<span>No Events Annotated</span>
			</div>
		{/each}
	</div>
</div>