<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Command from "$components/ui/command";
	import * as Popover from "$components/ui/popover";
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import { useSystemMapViewController } from "./controller.svelte";

	const view = useSystemMapViewController();
</script>

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
						<Button variant="outline" size="sm" onclick={view.retrySearch}>Retry search</Button>
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
