<script lang="ts">
	import type { Event, EventAnnotation } from "$lib/api";
	import { mdiPin, mdiPinOutline, mdiChatPlus, mdiMenuDown } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { useAnnotationDialogState } from "./annotation-dialog/dialogState.svelte";
	import { getEventKindIcon } from "./events";
	import EventTimeDate from "./EventTimeDate.svelte";

	type Props = {
		event: Event;
		annotations?: EventAnnotation[];
		pinned?: boolean;
		togglePinned?: () => void;
		loadingId?: string;
	}
	const { event, annotations = [], pinned, togglePinned, loadingId }: Props = $props();

	const annoDialog = useAnnotationDialogState();

	const canCreate = $derived(annoDialog.allowCreating);

	const attrs = $derived(event.attributes);

	const loading = $derived(!!loadingId && loadingId === event.id);
	const disabled = $derived(!!loadingId);

	const kindIcon = $derived(getEventKindIcon(attrs.kind));
</script>

{#snippet annotationBox(anno: EventAnnotation)}
	<div class="inline-block">
		<button onclick={() => annoDialog.setOpen(event, anno)} 
			class="max-w-32 min-w-12 h-fit border hover:border-neutral rounded p-1 bg-neutral-700/70 hover:bg-neutral-700/60 text-sm flex gap-2 flex-col cursor-pointer">
			<div class="flex gap-1 justify-between">
				<Avatar kind="user" id={anno.attributes.creator.id} size={14} />
				<Icon data={mdiMenuDown} size={14} />
			</div>
			<div class="text-neutral-content/80 leading-none text-start truncate w-full">{anno.attributes.notes}</div>
		</button>
	</div>
{/snippet}

<div class="h-[70px] group grid grid-cols-[80px_minmax(100px,1fr)_minmax(0,.4fr)] gap-2 place-items-center border py-1 px-2 bg-neutral-900/40 border-neutral-content/10 shadow-sm hover:shadow-md transition-shadow">
	<EventTimeDate timestamp={attrs.timestamp} />

	<div class="flex flex-col gap-1 w-full h-full justify-center items-start">
		<div class="flex gap-1 items-center">
			<Icon data={kindIcon.icon} classes={{ root: `rounded-full size-4 w-auto ${kindIcon.color}` }} />
			<span class="text-xs uppercase font-normal text-surface-content/50">{attrs.kind}</span>
		</div>
		<a href="/events/{event.id}" class="anchor link w-full truncate text-left align-baseline">{attrs.title}</a>
	</div>

	<div class="flex w-full h-full items-center justify-end gap-2">
		<div class="flex-1 h-full items-center justify-end flex gap-2">
			{#each annotations as anno}
				{@render annotationBox(anno)}
			{/each}

			{#if canCreate}
				<div class="hidden group-hover:inline w-fit h-full">
					<Button {disabled} onclick={() => annoDialog.setOpen(event)}>
						Annotate
						<Icon data={mdiChatPlus} />
					</Button>
				</div>
			{/if}
		</div>

		{#if !!togglePinned}
			<!--Tooltip title="Toggle Pinned">
				<Button iconOnly icon={pinned ? mdiPin : mdiPinOutline} {disabled} {loading} onclick={togglePinned} />
			</Tooltip-->
		{/if}
	</div>
</div>