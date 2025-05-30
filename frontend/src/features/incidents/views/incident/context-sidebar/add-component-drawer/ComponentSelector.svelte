<script lang="ts">
	import { Button, Checkbox, Icon, ListItem } from "svelte-ux";
	import { cls } from '@layerstack/tailwind';
	import { listSystemComponentsOptions, type SystemComponent } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiPlus } from "@mdi/js";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

	let showFilters = $state(false);

	const componentsQuery = createQuery(() =>
		listSystemComponentsOptions({
			query: {},
		})
	);
	const components = $derived(componentsQuery.data?.data ?? []);

	const setCreating = () => {
		alert("create new");
	}

	let selectedId = $state<string>();
</script>

<div class="flex flex-col h-full">
	<div class="p-2">filters</div>

	<div class="flex-1 flex flex-col min-h-0 overflow-y-auto">
		<LoadingQueryWrapper query={componentsQuery} view={componentsListView} />
	</div>
</div>

{#snippet componentsListView(components: SystemComponent[])}
	{#if components.length === 0}
		<div class="flex flex-col gap-2 py-4 rounded w-fit mx-auto">
			<span>No Components Found</span>
			<Button on:click={setCreating} color="secondary">
				Create Component
				<Icon data={mdiPlus} />
			</Button>
		</div>
	{/if}

	<div class="grid gap-4 bg-surface-200 p-1" class:hidden={components.length === 0}>
		{#each components as cmp (cmp.id)}
			<div>
				<ListItem
					title={cmp.attributes.name}
					subheading={cmp.attributes.description}
					on:click={() => {}}
					class={cls(
						"px-8 py-4",
						"cursor-pointer transition-shadow duration-100",
						"hover:bg-surface-100 hover:outline",
						selectedId == cmp.id ? "bg-surface-100 shadow-md" : ""
					)}
					noBackground
					noShadow
				>
					<div slot="actions">
						<Checkbox circle dense checked={selectedId == cmp.id} />
					</div>
				</ListItem>
			</div>
		{/each}
	</div>
{/snippet}
