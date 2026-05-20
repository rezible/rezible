<script lang="ts" module>
	import type { IncidentTimelineEvent } from "$lib/api";
	export type Props = {
		event?: IncidentTimelineEvent;
		selected: boolean;
	};
</script>

<script lang="ts">
	const {event, selected}: Props = $props();
</script>

<div class="p-2">
	<span>{event?.attributes.title || "title"}</span>
</div>
