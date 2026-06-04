<script lang="ts">
	import { Button } from "$components/ui/button";
	import Icon from "$components/common/icon/Icon.svelte";
	import { listSystemTopologyEntitiesOptions, type SystemTopologyEntity } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiPlus } from "@mdi/js";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";

	const entitiesQuery = createQuery(() =>
		listSystemTopologyEntitiesOptions({
			query: {},
		})
	);

	const setCreating = () => {
		alert("create new");
	}
</script>

<div class="flex flex-col h-full">
	<div class="p-2">filters</div>

	<div class="flex-1 flex flex-col min-h-0 overflow-y-auto">
		<LoadingQueryWrapper query={entitiesQuery} view={entitiesListView} />
	</div>
</div>

{#snippet entitiesListView(entities: SystemTopologyEntity[])}
	{#if entities.length === 0}
		<div class="flex flex-col gap-2 py-4 rounded w-fit mx-auto">
			<span>No topology entities found</span>
			<Button onclick={setCreating} color="secondary">
				Create Entity
				<Icon data={mdiPlus} />
			</Button>
		</div>
	{/if}

	<div class="grid gap-4 bg-surface-200 p-1" class:hidden={entities.length === 0}>
		{#each entities as entity (entity.id)}
			<div>
				<span>{entity.attributes.displayName}</span>
			</div>
		{/each}
	</div>
{/snippet}
