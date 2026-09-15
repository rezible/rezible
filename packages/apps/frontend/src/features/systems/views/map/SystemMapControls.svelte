<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Command from "$components/ui/command";
	import * as Popover from "$components/ui/popover";
	import * as Select from "$components/ui/select";
	import * as ToggleGroup from "$components/ui/toggle-group";
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import RiListView from "remixicon-svelte/icons/list-view";
	import RiPieChartLine from "remixicon-svelte/icons/pie-chart-line";
	import { useSystemMapViewController } from "./controller.svelte";

	const view = useSystemMapViewController();
</script>

<header class="border-border flex flex-wrap items-center justify-between gap-2 border-b px-4 py-3">
	<div class="flex min-w-0 flex-wrap items-center gap-2">
		<Popover.Root bind:open={view.searchOpen}>
			<Popover.Trigger>
				{#snippet child({ props })}
					<Button {...props} variant="outline" class="w-52 max-w-full justify-start">
						<RiSearchLine data-icon="inline-start" />
						<span class="text-muted-foreground">Find an entity…</span>
					</Button>
				{/snippet}
			</Popover.Trigger>
			<Popover.Content align="start" class="w-96 max-w-[calc(100vw-2rem)] p-0">
				<Command.Root shouldFilter={false}>
					<Command.Input bind:value={view.search} placeholder="Search names, kinds, or aliases…" />
					<Command.List class="max-h-80">
						{#if view.search.trim().length < 2}
							<Command.Empty>Enter at least two characters.</Command.Empty>
						{:else if view.searching}
							<Command.Loading>Searching…</Command.Loading>
						{:else if view.searchError}
							<div role="alert" class="flex flex-col gap-2 p-3">
								<p class="text-sm text-destructive">
									{view.searchError.detail || view.searchError.title}
								</p>
								<Button variant="outline" size="sm" onclick={view.retrySearch}
									>Retry search</Button
								>
							</div>
						{:else if view.searchResults.length === 0}
							<Command.Empty>No entities found.</Command.Empty>
						{:else}
							<Command.Group heading="Knowledge entities">
								{#each view.searchResults as entity (entity.id)}
									<Command.Item value={entity.id} onclick={() => view.focus(entity)}>
										<div class="min-w-0">
											<div class="truncate text-sm font-medium">
												{entity.attributes.latestState?.displayName ||
													entity.attributes.aliases[0]?.attributes.resourceRef
														.resourceRef ||
													"Unknown entity"}
											</div>
											<div class="text-muted-foreground truncate text-xs">
												{entity.attributes.kind.replaceAll("_", " ")}
											</div>
										</div>
									</Command.Item>
								{/each}
							</Command.Group>
						{/if}
					</Command.List>
				</Command.Root>
			</Popover.Content>
		</Popover.Root>
		<Select.Root
			type="single"
			value={String(view.depth)}
			onValueChange={(value) => {
				if (value) {
					view.setDepth(Number(value));
				}
			}}
		>
			<Select.Trigger class="w-36" aria-label="Neighborhood depth">
				Depth {view.depth}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each [1, 2, 3, 4] as depthOption (depthOption)}
						<Select.Item value={String(depthOption)}>Depth {depthOption}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
		<Select.Root
			type="single"
			value={view.kindFilter === "" ? "all" : view.kindFilter}
			onValueChange={(value) => view.setKindFilter(value === "all" ? "" : (value ?? ""))}
		>
			<Select.Trigger class="w-44" aria-label="Filter by subject type">
				{view.kindFilter === "" ? "All subject types" : view.kindFilter.replaceAll("_", " ")}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					<Select.Item value="all">All subject types</Select.Item>
					{#each view.availableKinds as kindOption (kindOption)}
						<Select.Item value={kindOption}>{kindOption.replaceAll("_", " ")}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
		<Select.Root
			type="single"
			value={view.predicateFilter === "" ? "all" : view.predicateFilter}
			onValueChange={(value) => view.setPredicateFilter(value === "all" ? "" : (value ?? ""))}
		>
			<Select.Trigger class="w-44" aria-label="Filter by relationship type">
				{view.predicateFilter === ""
					? "All relationships"
					: view.predicateFilter.replaceAll("_", " ")}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					<Select.Item value="all">All relationships</Select.Item>
					{#each view.availablePredicates as predicateOption (predicateOption)}
						<Select.Item value={predicateOption}>
							{predicateOption.replaceAll("_", " ")}
						</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
		{#if view.hasFilters}
			<Button variant="ghost" size="sm" onclick={() => view.clearFilters()}>Clear filters</Button>
		{/if}
	</div>
	<ToggleGroup.Root
		type="single"
		value={view.displayMode}
		onValueChange={(value) => {
			if (value === "graph" || value === "list") {
				view.setDisplayMode(value);
			}
		}}
		aria-label="Map display"
	>
		<ToggleGroup.Item value="graph">
			<RiPieChartLine />
			Graph
		</ToggleGroup.Item>
		<ToggleGroup.Item value="list">
			<RiListView />
			List
		</ToggleGroup.Item>
	</ToggleGroup.Root>
</header>
