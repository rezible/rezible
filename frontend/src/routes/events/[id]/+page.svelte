<script lang="ts">
	import { page } from "$app/state";
	import EventView from "$features/event/view/EventView.svelte";
	
</script>

<EventView id={page.params.id} />