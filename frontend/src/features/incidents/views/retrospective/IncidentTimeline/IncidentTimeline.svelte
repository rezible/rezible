<script lang="ts">
	import {
		cls,
		settings,
		Button,
		Icon,
		PeriodType,
		Timeline,
		TimelineEvent,
		Dialog,
		Header,
        Collapse
	} from 'svelte-ux';
	import {
		mdiAlarmLight,
		mdiCircleMedium,
		mdiClock,
		mdiClockOutline,
		mdiClose,
		mdiFire,
		mdiMagnify,
		mdiPencil,
		mdiPlus
	} from '@mdi/js';
	import type { Incident } from '$lib/api';
	import ConfirmChangeButtons from '$components/confirm-buttons/ConfirmButtons.svelte';
	import { events, type EventType, type IncidentEvent } from './events';
	import EventDetailsEdit from './EventEditor.svelte';
	import TimelineSummary from './TimelineSummary.svelte';

	interface Props { 
		incident: Incident;
	};
	let { incident }: Props = $props();

	const timelineElements: { [i: number]: HTMLElement } = {};

	type EditMode = 'edit' | 'create';
	type EditState = {
		event: IncidentEvent;
		mode: EditMode;
		changed: boolean;
	};

	let editorOpen = $state(false);
	let eventsOpen = $state(false);

	let editing = $state<EditState>();
	let selected = $state<IncidentEvent>();
	let hoveringId = $state<string>();

	const getIcon = (eventType: EventType) => {
		// switch (eventType) {
		//     case 'impact': return mdiFire;
		//     case 'detection': return mdiAlarmLight;
		//     case 'investigation': return mdiMagnify;
		// }
		return mdiCircleMedium;
	};

	const getDialogTitle = (mode?: EditMode) => {
		if (!mode) return '';
		if (mode === 'create') return 'Create Timeline Event';
		return `Edit Timeline Event`;
	};
	const dialogTitle = $derived(getDialogTitle(editing?.mode));

	const createEvent = () => {
		// TODO: base this on last event?
		const newEvent: IncidentEvent = {
			id: '',
			title: "",
			start: new Date(),
			type: 'note',
			description: "",
			data: []
		}
		editing = {event: newEvent, mode: 'create', changed: false};
		editorOpen = true;
	};

	const editEvent = (event: IncidentEvent) => {
		editing = { event, mode: 'edit', changed: false };
		editorOpen = true;
	};

	const stopEditing = () => {
		editing = undefined;
		editorOpen = false;
	};

	const saveChanges = () => {
		if (!editing) return;
		stopEditing();
	};

	const summaryEventClicked = (id: string) => {
		selected = events.find(e => e.id === id);
	}

	const eventHover = (event: IncidentEvent, hovering: boolean) => {
		if (!hovering && hoveringId === event.id) hoveringId = undefined;
		if (hovering && hoveringId !== event.id) hoveringId = event.id;
	};
</script>


<div class="flex h-10">
	<div class="flex-1 flex h-10 items-end">
		<span class="text-lg text-surface-content/80">Incident Timeline</span>
	</div>
	<div class="">
		<Button variant="fill-light" rounded={false} on:click={createEvent}>
			Add Event
			<Icon data={mdiPlus} />
		</Button>
	</div>
</div>

<div class="border border-surface-content/15 bg-surface-content/5 p-2 px-3 rounded-lg mt-1">
	<TimelineSummary
		{events}
		selectedId={selected?.id}
		{hoveringId}
		onEventClicked={summaryEventClicked}
	/>

	<Collapse
		bind:open={eventsOpen}
		class="bg-surface-100 elevation-1 first:rounded-t last:rounded-b"
	>
		<div slot="trigger" class="flex-1 px-3 py-3">Events</div>
		<div class="">
			{@render eventsList()}
		</div>
	</Collapse>

	<Dialog
		bind:open={editorOpen}
		persistent
		portal
		classes={{ root:"p-8", dialog: 'flex flex-col max-w-full max-w-4xl h-full' }}
	>
		<div slot="header" class="border-b p-2">
			<span class="text-xl">{dialogTitle}</span>
		</div>

		{#if editing}
			<EventDetailsEdit {incident} event={editing.event} bind:changed={editing.changed} />
		{/if}

		<svelte:fragment slot="actions">
			<ConfirmChangeButtons
				onClose={stopEditing}
				onConfirm={saveChanges}
				saveEnabled={editing?.changed}
			/>
		</svelte:fragment>
	</Dialog>
</div>

{#snippet eventsList()}
<div class="w-full flex gap-1">
	<div class="overflow-y-auto max-h-72 bg-surface-100 flex-1 flex flex-col divide-y">
		{#each events as e, i (i)}
			<div
				class="p-2 cursor-pointer {e.id === selected?.id
					? 'bg-secondary-800'
					: 'hover:bg-secondary-900'}"
				role="none"
				onmouseenter={() => {
					eventHover(e, true);
				}}
				onmouseleave={() => {
					eventHover(e, false);
				}}
				onclick={() => {
					selected = e;
				}}
			>
				{e.title}
			</div>
		{/each}
	</div>
	{#if selected}
		<div class="border-2 w-1/2 p-2">
			<Header title="Event Details" class="">
				<div slot="actions">
					<Button
						icon={mdiPencil}
						on:click={() => {
							if (selected) editEvent($state.snapshot(selected));
						}}
					>
						Edit
					</Button>
					<Button
						iconOnly
						icon={mdiClose}
						on:click={() => {
							selected = undefined;
						}}
					/>
				</div>
			</Header>
			foo
		</div>
	{/if}
</div>
{/snippet}