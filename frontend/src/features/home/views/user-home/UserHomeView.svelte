<script lang="ts">
	import ShiftsHeader from "$features/home/views/user-home/shifts-header/ShiftsHeader.svelte";
	import UserItems from "$features/home/views/user-home/user-items/UserItems.svelte";
	import UserPinnedItems from "$features/home/views/user-home/user-items/UserPinnedItems.svelte";
	import OncallEventsTable from "$features/home/views/user-home/oncall-events-table/OncallEventsTable.svelte";
</script>

<div class="h-full w-full flex flex-col gap-2">
	<ShiftsHeader />

	<div class="flex-1 flex min-h-0 gap-2">
		<div class="flex-1">
			<OncallEventsTable />
		</div>
		<div class="w-2/5 flex flex-col gap-2 min-h-0">
			<UserItems />
			<UserPinnedItems />
		</div>
	</div>
</div>
