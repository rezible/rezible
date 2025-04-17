<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	
	type Props = {
		event: OncallEvent;
		current?: OncallAnnotation;
	}
	const { event, current }: Props = $props();

	const attrs = $derived(event.attributes);
</script>

{#if attrs.kind === "alert"}
	<span>alert event {event.id}</span>
{/if}