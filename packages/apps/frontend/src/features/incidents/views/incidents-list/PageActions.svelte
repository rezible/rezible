<script lang="ts">
	import { mdiPlus } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import { Button } from "$components/ui/button";
	import { useIncidentCreateDialog } from "$features/incidents/components/create-incident-dialog";

	const incidentDialog = useIncidentCreateDialog();
</script>

<div class="flex gap-2">
	<Button onclick={() => {incidentDialog.openCreate()}}>
		Create Incident
		<Icon data={mdiPlus} />
	</Button>
</div>
