<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	
	type Props = {
		event: OncallEvent;
		current?: OncallAnnotation;
	}
	const { event, current }: Props = $props();

	const attrs = $derived(event.attributes);
</script>

<div class:hidden={attrs.kind !== "alert"}>
	<span>alert event {event.id}</span>
</div>

<div class="flex flex-col gap-2">
	<span>{attrs.kind}</span>
</div>