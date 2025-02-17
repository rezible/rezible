<script lang="ts">
	import { Button, Checkbox, cls, Header, Icon, ListItem } from "svelte-ux";
	import { listSystemComponentsOptions, type SystemComponent } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiFilter, mdiPlus } from "@mdi/js";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import { componentDialog } from "./dialogState.svelte";

	let showFilters = $state(false);

	const componentsQuery = createQuery(() =>
		listSystemComponentsOptions({
			query: {},
		})
	);
	const components = $derived(componentsQuery.data?.data ?? []);

	const selectedId = $derived(componentDialog.selectedAddComponent?.id);
</script>

<div class="flex flex-col gap-2 p-2 border rounded-lg">
	<div class="w-full border-b pb-2">
		<Header title="Components">
			<svelte:fragment slot="actions">
				<Button
					icon={mdiFilter}
					on:click={() => {
						showFilters = !showFilters;
					}}
				>
					{showFilters ? "Hide" : "Show"} Filters
				</Button>

				{#if components.length > 0}
					<Button on:click={componentDialog.setCreating} color="secondary">
						Create New
						<Icon data={mdiPlus} />
					</Button>
				{/if}
			</svelte:fragment>
		</Header>

		<div class="" class:hidden={!showFilters}>filters</div>
	</div>

	<LoadingQueryWrapper query={componentsQuery} view={componentsListView} />
</div>

{#snippet componentsListView(components: SystemComponent[])}
	{#if components.length === 0}
		<div class="flex flex-col gap-2 py-4 rounded w-fit mx-auto">
			<span>No Components Found</span>
			<Button on:click={componentDialog.setCreating} color="secondary">
				Create Component
				<Icon data={mdiPlus} />
			</Button>
		</div>
	{/if}

	<div class="grid gap-4 bg-surface-200 p-4" class:hidden={components.length === 0}>
		{#each components as cmp (cmp.id)}
			<div>
				<ListItem
					title={cmp.attributes.name}
					subheading={cmp.attributes.description}
					on:click={() => {componentDialog.setSelectedAddComponent(cmp)}}
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
