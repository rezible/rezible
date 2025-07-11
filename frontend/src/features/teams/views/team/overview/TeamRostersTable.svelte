<script lang="ts">
	import Avatar from "$src/components/avatar/Avatar.svelte";
	import Card from "$src/components/card/Card.svelte";
	import Header from "$src/components/header/Header.svelte";
	import { listOncallRostersOptions, type OncallRoster } from "$src/lib/api";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { Pagination, Table } from "svelte-ux";
	import { useTeamViewState } from "../viewState.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import { tableCell, type ColumnDef } from "@layerstack/svelte-table";
	import { fromStore } from "svelte/store";
	import { watch } from "runed";

	const viewState = useTeamViewState();
	const teamId = $derived(viewState.teamId);

	const columns: ColumnDef<OncallRoster>[] = [
		{
			name: "avatar",
			header: "",
			classes: {th: "w-9"}
		},
		{
			name: "name",
			header: "",
			value: data => data.attributes.name,
		},
		{
			name: "actions",
			header: "",
			align: "right",
		},
	];

	const paginationStore = createPaginationStore({ page: 1, perPage: 10, total: 0 });
	const pagination = fromStore(paginationStore);

	const rostersQuery = createQuery(() => ({ ...listOncallRostersOptions({ query: { teamId } }) }));
	const rosters = $derived(rostersQuery.data?.data ?? []);
	const numRosters = $derived(rosters.length);
	watch(() => numRosters, total => {paginationStore.setTotal(total)});

	const rostersPage = $derived(rosters);
</script>

{#snippet rostersTableView()}
	<Table data={rostersPage} {columns}>
		<tbody slot="data" let:columns let:data let:getCellValue>
			{#each data ?? [] as rowData, rowIndex}
				<tr class="">
					{#each columns as column (column.name)}
						{@const value = getCellValue(column, rowData, rowIndex)}

						<td use:tableCell={{ column, rowData, rowIndex, tableData: data }} class="w-fit">
							{#if column.name === "avatar"}
								<Avatar kind="roster" size={24} id={rowData.id} />
							{:else if column.name === "actions"}
								<span>edit</span>
							{:else}
								{value}
							{/if}
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</Table>
	<Pagination
		pagination={paginationStore}
		perPageOptions={[5, 10, 25, 100]}
		show={["perPage", "pagination", "prevPage", "nextPage"]}
		classes={{
			root: "border-t py-1 mt-2",
			perPage: "flex-1 text-right",
			pagination: "px-8",
		}}
	/>
{/snippet}

<Card classes={{ root: "max-w-lg", headerContainer: "p-3" }} contents={rostersTableView}>
	{#snippet header()}
		<Header title="Rosters" classes={{title: "text-xl"}} />
	{/snippet}
</Card>
