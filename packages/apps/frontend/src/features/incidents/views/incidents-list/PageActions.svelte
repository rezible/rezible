<script lang="ts">
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import { Button } from "$components/ui/button";

	// import { useIncidentCreateDialog } from "$features/incidents/components/create-incident-dialog";
	// const incidentDialog = useIncidentCreateDialog();
</script>

<div class="flex gap-2">
	<Button
		onclick={() => {
			// incidentDialog.openCreate();
		}}
	>
		Create Incident
		<RiAddLine class="" aria-hidden="true" />
	</Button>
</div>
