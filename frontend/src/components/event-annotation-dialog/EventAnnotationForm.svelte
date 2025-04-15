<script lang="ts">
	import type { OncallEvent } from "$lib/api";
	
	type Props = {
		event: OncallEvent;
	}
	const { event }: Props = $props();
</script>

<span>event {event.id}</span>