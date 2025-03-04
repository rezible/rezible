<script lang="ts" module>
	import type { IncidentEvent } from "$lib/api";
	export type TimelineEventComponentProps = {
		event: IncidentEvent;
	};
</script>

<script lang="ts">
	let { event }: TimelineEventComponentProps = $props();
</script>

<span>{event.attributes.title}</span>
