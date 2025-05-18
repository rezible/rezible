<script lang="ts" module>
	import type { IncidentEvent } from "$lib/api";
	export type Props = {
		event?: IncidentEvent;
		selected: boolean;
	};
</script>

<script lang="ts">
	const {event, selected}: Props = $props();
</script>

<div class="p-2">
	<span>{event?.attributes.title || "title"} - selected:${selected}</span>
</div>
