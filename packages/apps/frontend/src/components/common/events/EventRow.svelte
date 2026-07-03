<script lang="ts">
	import type { Event, EventAnnotation } from "$lib/api";
	import { mdiMenuDown } from "@mdi/js";
	import Icon from "$components/common/icon/Icon.svelte";
	import Avatar from "$components/common/entity-avatar/EntityAvatar.svelte";
	import { formatDate } from "date-fns";

	type Props = {
		event: Event;
		annotations?: EventAnnotation[];
	}
	const { event, annotations = [] }: Props = $props();

	const openAnnotationDialog = (event?: Event, anno?: EventAnnotation) => {
		 alert("open annotation dialog");
	}

	const attrs = $derived(event.attributes);

	const date = $derived(new Date(attrs.occurredAt));
	const humanDate = $derived(formatDate(date, 'MMM d'));
</script>

{#snippet annotationBox(anno: EventAnnotation)}
	<div class="inline-block">
		<button onclick={() => openAnnotationDialog()} 
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
	<div class="flex flex-col gap-1 justify-between w-full items-start">
		<span class="text-sm flex items-center gap-1">
			{humanDate}
		</span>
	</div>

	<div class="flex flex-col gap-1 w-full h-full justify-center items-start">
		<div class="flex gap-1 items-center">
			<span class="text-xs uppercase font-normal text-surface-content/50">{attrs.kind}</span>
		</div>
		<a href="/events/{event.id}" class="anchor link w-full truncate text-left align-baseline">{attrs.provider} {attrs.providerSubjectRef}</a>
	</div>

	<div class="flex w-full h-full items-center justify-end gap-2">
		<div class="flex-1 h-full items-center justify-end flex gap-2">
			{#each annotations as anno}
				{@render annotationBox(anno)}
			{/each}
		</div>
	</div>
</div>