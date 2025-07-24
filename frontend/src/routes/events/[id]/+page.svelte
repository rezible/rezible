<script lang="ts">
	import { page } from "$app/state";
	import EventView from "$features/events/views/event/EventView.svelte";
	
</script>

<EventView id={page.params.id} />