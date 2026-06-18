<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";

	import EventsTable from "./events-table/EventsTable.svelte";
	import UserBacklogBox from "./user-backlog/UserBacklogBox.svelte";

	setPageBreadcrumbs(() => [{ label: "Home" }]);
</script>

<div class="flex-1 flex min-h-0 gap-2">
	<!-- <div class="flex-1 flex flex-col gap-2">
		<EventsTable />
	</div>

	<div class="w-2/5 flex flex-col gap-2 min-h-0">
		<UserBacklogBox />
	</div> -->
</div>