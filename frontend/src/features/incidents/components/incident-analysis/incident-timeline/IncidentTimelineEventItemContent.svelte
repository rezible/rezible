<script lang="ts" module>
	import type { IncidentEvent } from "$lib/api";
	export type Props = {
		event: IncidentEvent;
	};
</script>

<script lang="ts">
	let { event }: Props = $props();
</script>

<span>{event.attributes.title}</span>
