<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import ListRow from "$components/common/list-row/ListRow.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { shiftStatus } from "$features/oncall/lib/shift-status";
	import { initOncallShiftsListController } from "./controller.svelte";

	registerPageDescriptor(() => ({ title: "Oncall Shifts" }));

	const controller = initOncallShiftsListController();
</script>

<FilterPage title="Shifts" count={controller.total} updating={controller.query.isFetching}>
	<PaginatedQueryListBox {...controller.paginatedShiftsQuery} showUpdating={false}>
		<LoadingQueryWrapper
			query={controller.query}
			feedback="quiet"
			isEmpty={(shifts) => shifts.length === 0}
		>
			{#snippet loading()}
				<ListSkeleton label="Loading shifts" />
			{/snippet}
			{#snippet empty()}
				<p class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">No shifts found.</p>
			{/snippet}
			{#snippet view(shifts)}
				<ul class="divide-y overflow-hidden rounded-lg border bg-card">
					{#each shifts as shift (shift.id)}
						{#snippet window()}
							<Timestamp value={shift.attributes.startAt} format="absolute" />
							<span aria-hidden="true">–</span>
							<Timestamp value={shift.attributes.endAt} format="absolute" />
						{/snippet}
						<ListRow
							title={shift.attributes.roster.attributes.name}
							href={resolve("/(hidden)/oncall/shifts/[id]", { id: shift.id })}
							summary={shift.attributes.user.attributes.name}
							status={shiftStatus(shift, controller.now)}
							facts={window}
						/>
					{/each}
				</ul>
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
