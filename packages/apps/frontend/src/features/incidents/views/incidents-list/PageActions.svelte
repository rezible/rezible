<script lang="ts">
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import { Button } from "$components/ui/button";

	import { useIncidentCreateDialog } from "$features/incidents/components/create-incident-dialog";

	const incidentDialog = useIncidentCreateDialog();
</script>

{#if incidentDialog.enabled}
	<Button onclick={() => incidentDialog.openCreate()}>
		<RiAddLine aria-hidden="true" />
		Declare incident
	</Button>
{/if}
