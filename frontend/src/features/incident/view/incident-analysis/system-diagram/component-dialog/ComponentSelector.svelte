<script lang="ts">
	import { Checkbox, ListItem } from "svelte-ux";
	import Button from "$components/button/Button.svelte";
	import { cls } from '@layerstack/tailwind';
	import Icon from "$components/icon/Icon.svelte";
	import { listSystemComponentsOptions, type SystemComponent } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiPlus } from "@mdi/js";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { useComponentDialog } from "./dialogState.svelte";

	const dialog = useComponentDialog();

	let showFilters = $state(false);

	const componentsQuery = createQuery(() =>
		listSystemComponentsOptions({
			query: {},
		})
	);
	const components = $derived(componentsQuery.data?.data ?? []);

	const selectedId = $derived(dialog.selectedAddComponent?.id);
</script>

<div class="flex flex-col gap-2">
	<div class="flex w-full justify-between items-center border-b pb-2">
		<div class="">filters</div>

		{#if components.length > 0}
			<Button on:click={() => dialog.setCreating()} color="secondary">
				Create New
				<Icon data={mdiPlus} />
			</Button>
		{/if}
	</div>

	<LoadingQueryWrapper query={componentsQuery} view={componentsListView} />
</div>

{#snippet componentsListView(components: SystemComponent[])}
	{#if components.length === 0}
		<div class="flex flex-col gap-2 py-4 rounded w-fit mx-auto">
			<span>No Components Found</span>
			<Button on:click={() => dialog.setCreating()} color="secondary">
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
					on:click={() => {dialog.setSelectedAddComponent(cmp)}}
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
