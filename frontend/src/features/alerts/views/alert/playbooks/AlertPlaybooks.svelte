<script lang="ts">
	import { useAlertViewController } from "$features/alerts/views/alert";
	import { listPlaybooksOptions, type ListPlaybooksData, type Playbook } from "$src/lib/api";
	import { QueryPaginatorState } from "$src/lib/paginator.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import RosterSelectField from "$src/components/roster-select-field/RosterSelectField.svelte";

	const view = useAlertViewController();

	const paginator = new QueryPaginatorState();

	let rosterId = $state<string>();
	const onRosterSelected = (id?: string) => (rosterId = id);

	const queryParams = $derived<ListPlaybooksData["query"]>({
		alertId: view.alertId,
		...paginator.queryParams,
	});
	const query = createQuery(() => listPlaybooksOptions({ query: queryParams }));
	paginator.watchQuery(query);
</script>

{#snippet playbookListItem(pb: Playbook)}
	<a href="/playbooks/{pb.id}">
		<span>{pb.attributes.title}</span>
	</a>
{/snippet}

<div class="w-full h-full flex flex-col gap-2">
	<div class="flex gap-2">
		<RosterSelectField onSelected={onRosterSelected} selectedId={rosterId} />
	</div>

	<div class="flex-1 flex flex-col gap-1 border">
		<LoadingQueryWrapper {query}>
			{#snippet view(playbooks: Playbook[])}
				{#each playbooks as pb}
					{@render playbookListItem(pb)}
				{:else}
					<span>No results</span>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>

	<!-- <Pagination {...paginator.paginationProps} /> -->
</div>

