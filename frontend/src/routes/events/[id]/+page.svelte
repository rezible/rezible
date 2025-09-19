<script lang="ts">
	import { page } from "$app/state";
	import { EventView, setEventViewState } from "$src/features/event";

	setEventViewState(() => page.params.id);
</script>

<EventView />