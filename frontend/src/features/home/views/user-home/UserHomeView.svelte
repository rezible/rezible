<script lang="ts">
	import ShiftsHeader from "$features/home/views/user-home/shifts-header/ShiftsHeader.svelte";
	import OncallEventsTable from "$features/home/views/user-home/oncall-events-table/OncallEventsTable.svelte";
	import UserBacklogBox from "$features/home/views/user-home/user-backlog/UserBacklogBox.svelte";
	import UserRelevantBox from "$features/home/views/user-home/user-relevant/UserRelevantBox.svelte";
</script>

<div class="h-full w-full flex flex-col gap-2">
	<ShiftsHeader />

	<div class="flex-1 flex min-h-0 gap-2">
		<div class="flex-1 flex flex-col gap-2">
			<OncallEventsTable />
		</div>

		<div class="w-3/5 flex flex-col gap-2 min-h-0">
			<UserBacklogBox />

			<!-- <UserRelevantBox /> -->
		</div>
	</div>
</div>
