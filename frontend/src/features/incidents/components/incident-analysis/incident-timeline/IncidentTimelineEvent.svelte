<script lang="ts" module>
	import type { IncidentEvent } from "$lib/api";
	export type TimelineEventComponentProps = {
		event: IncidentEvent;
	};
</script>

<script lang="ts">
	import { onMount } from "svelte";

	let { event }: TimelineEventComponentProps = $props();

	onMount(() => {
		// console.log("mounted");
		// return () => {console.log("unmounting")}
	});
</script>

<span>{event.id}</span>
