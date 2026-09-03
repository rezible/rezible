<script lang="ts">
	import { resolve } from "$app/paths";
	import { useAlertViewController } from "$features/alerts/views/alert";
	import { listPlaybooksOptions, type ListPlaybooksData, type Playbook } from "$lib/api";
	import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import RosterSelectField from "$components/forms/roster-select-field/RosterSelectField.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";

	const controller = useAlertViewController();

	let rosterId = $state<string>();
	const onRosterSelected = (id?: string) => (rosterId = id);

	const queryParams = $derived<ListPlaybooksData["query"]>({
		alertId: controller.alertId,
	});
	const paginatedPlaybooksQuery = createPaginatedQuery({
		queryOptions: (pagination) => listPlaybooksOptions({ query: {...queryParams, ...pagination} }),
	});
</script>

{#snippet playbookListItem(pb: Playbook)}
	<a href={resolve(`/playbooks/${pb.id}`)}>
		<span>{pb.attributes.title}</span>
	</a>
{/snippet}

<div class="w-full h-full flex flex-col gap-2">
	<div class="flex gap-2">
		<RosterSelectField onSelected={onRosterSelected} selectedId={rosterId} />
	</div>

	<div class="flex-1 min-h-0 border p-1">
		<PaginatedQueryListBox {...paginatedPlaybooksQuery}>
			<LoadingQueryWrapper query={paginatedPlaybooksQuery.query}>
				{#snippet view(playbooks: Playbook[])}
					{#each playbooks as pb (pb.id)}
						{@render playbookListItem(pb)}
					{:else}
						<span>No results</span>
					{/each}
				{/snippet}
			</LoadingQueryWrapper>
		</PaginatedQueryListBox>
	</div>
</div>
