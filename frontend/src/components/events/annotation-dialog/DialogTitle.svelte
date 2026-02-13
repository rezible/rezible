<script lang="ts">
	import Icon from "$components/icon/Icon.svelte";
	import type { Event } from "$lib/api";
	import { mdiCalendar, mdiClose, mdiPhoneAlert } from "@mdi/js";
	import { useAnnotationDialogState } from "./dialogState.svelte";
	import { Button } from "$components/ui/button";
	import { getEventKindIcon, getEventTimeIcon } from "../events";
	import EventTimeDate from "../EventTimeDate.svelte";


	type Props = {
		event: Event;
		close: () => void;
	};
	const { event, close }: Props = $props();

	const dialog = useAnnotationDialogState();

	const eventKind = $derived(event.attributes.kind === "alert" ? "Alert" : "Event");
	const title = $derived(dialog.view === "view" ? "Viewing Annotation" : `Annotating ${eventKind}`);

	const kindIcon = $derived(getEventKindIcon(event.attributes.kind));
</script>

<div class="flex p-2 pb-0">
	<div class="flex flex-col flex-1">
		<span class="text-surface-content/50">{title}</span>
		<h1 class="text-lg flex items-center gap-2">
			<Icon data={kindIcon.icon} classes={{ root: `rounded-full size-5 w-auto ${kindIcon.color}` }} />
			{event.attributes.title}
		</h1>
		<EventTimeDate timestamp={event.attributes.timestamp} />
	</div>
	<div class="">
		<Button onclick={close}>close</Button>
	</div>
</div>