<script lang="ts">
	import ShiftsHeader from "$features/home/components/shifts-header/ShiftsHeader.svelte";
	import UserItems from "$features/home/components/user-items/UserItems.svelte";
	import UserPinnedItems from "$features/home/components/user-items/UserPinnedItems.svelte";
	import EventsOverview from "$features/home/components/events-overview/EventsOverview.svelte";
</script>

<div class="h-full w-full flex flex-col gap-2">
	<ShiftsHeader />

	<div class="flex-1 flex min-h-0 gap-2">
		<div class="flex-1">
			<EventsOverview />
		</div>
		<div class="w-1/3 flex flex-col gap-2 min-h-0">
			<UserItems />
			<UserPinnedItems />
		</div>
	</div>
</div>
