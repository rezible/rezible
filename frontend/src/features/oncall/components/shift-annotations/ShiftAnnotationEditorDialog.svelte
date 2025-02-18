<script lang="ts">
	import {
		createOncallShiftAnnotationMutation,
		listOncallShiftAlertsOptions,
		listOncallShiftIncidentsOptions,
		type CreateOncallShiftAnnotationRequestBody,
	} from "$lib/api";
	import { mdiFilter, mdiClose, mdiChatPlus } from "@mdi/js";
	import { Icon, Button, TextField, Header, Dialog } from "svelte-ux";
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import {
		createTimeline,
		eventKindIcons,
		type ShiftTimelineEvent,
		type ShiftTimelineNode,
	} from "$features/oncall/lib/handover-timeline";

	type Props = {
		shiftId: string;
		annotatedEventIds: Set<string>;
		open: boolean;
		onCreated: VoidFunction;
	};
	let { shiftId, annotatedEventIds, open = $bindable(), onCreated }: Props = $props();

	const incidentsQuery = createQuery(() => ({
		...listOncallShiftIncidentsOptions({ path: { id: shiftId } }),
		enabled: open,
	}));
	const incidents = $derived(incidentsQuery.data?.data);

	const alertsQuery = createQuery(() => ({
		...listOncallShiftAlertsOptions({ path: { id: shiftId } }),
		enabled: open,
	}));
	const alerts = $derived(alertsQuery.data?.data);

	const timeline = $derived(createTimeline(annotatedEventIds, incidents, alerts));

	type DraftAnnotation = {
		event: ShiftTimelineEvent;
		notes: string;
		pinned: boolean;
	};
	let draftAnnotation = $state<DraftAnnotation>();

	const setAnnotationEvent = (e: ShiftTimelineEvent) => {
		draftAnnotation = {
			event: e,
			notes: "",
			pinned: false,
		};
	};

	const clearAnnotation = () => {
		draftAnnotation = undefined;
	};

	const createAnnotationMutation = createMutation(() => ({
		...createOncallShiftAnnotationMutation(),
		onSuccess: () => {
			onCreated();
			clearAnnotation();
		},
	}));

	const saveAnnotation = () => {
		if (!draftAnnotation) return;
		const d = $state.snapshot(draftAnnotation);
		const body: CreateOncallShiftAnnotationRequestBody = {
			attributes: {
				eventId: d.event.eventId,
				eventKind: d.event.kind,
				notes: d.notes,
				pinned: d.pinned,
				minutesOccupied: 0,
				occurredAt: d.event.occurredAt,
				title: d.event.title,
			},
		};
		createAnnotationMutation.mutate({ path: { id: shiftId }, body });
	};
</script>

<Dialog
	bind:open
	loading={createAnnotationMutation.isPending}
	persistent
	portal
	classes={{
		dialog: "flex flex-col max-h-full w-5/6 max-w-7xl my-2",
		root: "p-4",
	}}
>
	<div slot="header" class="border-b p-2" let:close>
		<Header title={!!draftAnnotation ? "Annotating Event" : "Shift Events"}>
			<svelte:fragment slot="actions">
				<Button on:click={() => close({ force: true })} iconOnly icon={mdiClose} />
			</svelte:fragment>
		</Header>
	</div>

	<svelte:fragment>
		{#if open}{@render dialogBody()}{/if}
	</svelte:fragment>
</Dialog>

{#snippet dialogBody()}
	<div class="flex flex-col gap-2 overflow-y-auto p-2">
		{#if draftAnnotation}
			<Header
				title={draftAnnotation.event.title}
				subheading={draftAnnotation.event.occurredAt.toLocaleString()}
			/>
			<div class="w-full border-t pt-2">
				<TextField
					label="Notes"
					multiline
					bind:value={draftAnnotation.notes}
					placeholder="Any notes on this event"
					classes={{ container: "bg-surface-300" }}
				/>
			</div>
			<div class="w-full flex justify-end">
				<ConfirmButtons
					closeText="Cancel"
					onClose={clearAnnotation}
					onConfirm={saveAnnotation}
					confirmText="Save"
				/>
			</div>
		{:else}
			<div class="border-b pb-2">
				<Button variant="fill-light">Filter <Icon data={mdiFilter} /></Button>
			</div>

			{#if timeline.length == 0}
				<span>No Events For Shift</span>
			{:else}
				<div class="min-h-0 flex-1 overflow-y-auto flex flex-col gap-4">
					{#each timeline as node, i}
						{@render timelineNode(node)}
					{/each}
				</div>
			{/if}
		{/if}
	</div>
{/snippet}

{#snippet timelineNode(node: ShiftTimelineNode)}
	{@const event = node.event}
	{@const icon = eventKindIcons[event.kind]}
	{@const isPinned = false}
	<div
		class="grid grid-cols-[100px_auto_minmax(0,1fr)] place-items-center border p-2 hover:bg-surface-100 bg-surface-200"
	>
		<div class="justify-self-start">
			<span class="flex items-center">
				{event.occurredAt.toLocaleString()}
			</span>
		</div>

		<div class="items-center static z-10">
			<Icon data={icon} classes={{ root: "bg-accent-900 rounded-full p-2 w-auto h-10" }} />
		</div>

		<div class="w-full justify-self-start grid grid-cols-[auto_40px] items-center px-2">
			<div class="leading-none">{event.title}</div>
			<div class="place-self-end">
				<Button
					icon={mdiChatPlus}
					iconOnly
					on:click={() => {
						setAnnotationEvent(event);
					}}
				/>
			</div>
		</div>
	</div>
{/snippet}
