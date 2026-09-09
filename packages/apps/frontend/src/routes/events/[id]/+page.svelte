<script lang="ts">
	import type { PageProps } from "./$types";
	import EventView from "$features/events/views/event";

	const { params }: PageProps = $props();
</script>

<EventView id={params.id} />
