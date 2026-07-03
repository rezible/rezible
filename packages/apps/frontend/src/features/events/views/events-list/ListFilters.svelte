<script lang="ts">
	import { useEventsListController } from "./controller.svelte";

	const controller = useEventsListController();
	const filters = $derived(controller.filters);
</script>

<div class="flex flex-col gap-2">

</div>
