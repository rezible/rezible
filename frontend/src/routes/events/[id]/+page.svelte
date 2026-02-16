<script lang="ts">
	import { page } from "$app/state";
	import EventView from "$features/events/views/event";
</script>

<EventView id={page.params.id || ""} />