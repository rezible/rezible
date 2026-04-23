<script lang="ts">
	import { mdiPlus } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import { useIncidentCreateDialog } from "$features/incidents/components/create-incident-dialog";

	import EventsTable from "./events-table/EventsTable.svelte";
	import UserBacklogBox from "./user-backlog/UserBacklogBox.svelte";

	setPageBreadcrumbs(() => [{ label: "Home" }]);

	const incidentDialog = useIncidentCreateDialog();
</script>

<div class="flex flex-col gap-2">
	<div class="flex items-center justify-end">
		<Button onclick={() => {incidentDialog.openCreate()}}>
			Create Incident
			<Icon data={mdiPlus} />
		</Button>
	</div>

	<div class="flex-1 flex min-h-0 gap-2">
		<div class="flex-1 flex flex-col gap-2">
			<EventsTable />
		</div>

		<div class="w-2/5 flex flex-col gap-2 min-h-0">
			<UserBacklogBox />
		</div>
	</div>
</div>
