<script lang="ts" module>
	import type { IncidentEvent } from "$lib/api";
	export type Props = {
		event: IncidentEvent;
		selected: boolean;
	};
</script>

<script lang="ts">
	let { event, selected }: Props = $props();
</script>

<span>{event.attributes.title} selected:{selected}</span>
