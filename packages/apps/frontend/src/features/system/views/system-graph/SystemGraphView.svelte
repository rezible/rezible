<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { QueryPaginatorState } from "$lib/paginator.svelte";
	import { listSystemTopologyEntitiesOptions, type ListSystemTopologyEntitiesData, type SystemTopologyEntity } from "$lib/api";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";

	setPageBreadcrumbs(() => [{ label: "System Graph" }]);

	const paginator = new QueryPaginatorState();
	let searchValue = $state<string>();
	const params = $derived<ListSystemTopologyEntitiesData["query"]>({
		search: searchValue,
		...paginator.queryParams,
	});
	const query = createQuery(() => listSystemTopologyEntitiesOptions({ query: params }));
	paginator.watchQuery(query);
</script>

<span>graph view</span>
