<script lang="ts">
	import { appShell } from "$features/app";

	import ShiftsHeader from "./shifts-header/ShiftsHeader.svelte";
	import EventsTable from "./events-table/EventsTable.svelte";
	import UserBacklogBox from "./user-backlog/UserBacklogBox.svelte";
	import UserRelevantBox from "./user-relevant/UserRelevantBox.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Home" }]);
</script>

<div class="h-full w-full flex flex-col gap-2">
	<ShiftsHeader />

	<div class="flex-1 flex min-h-0 gap-2">
		<div class="flex-1 flex flex-col gap-2">
			<EventsTable />
		</div>

		<div class="w-2/5 flex flex-col gap-2 min-h-0">
			<UserBacklogBox />

			<!-- <UserRelevantBox /> -->
		</div>
	</div>
</div>
