<script lang="ts">
	import { useIncidentView } from "$features/incidents/views/incident";

	const view = useIncidentView();

	const attrs = $derived(view.incident?.attributes);
</script>

<div class="flex flex-col gap-2 overflow-y-auto">

</div>
