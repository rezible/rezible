<script lang="ts">
	import * as Button from "$components/ui/button";
	import * as Command from "$components/ui/command";
	import * as Popover from "$components/ui/popover";
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import { useSystemMapViewController } from "./controller.svelte";

	const view = useSystemMapViewController();
</script>

<Popover.Root bind:open={view.searchOpen}>
	<Popover.Trigger>
		{#snippet child({ props })}
			<Button.Root {...props} variant="outline" class="bg-background w-72 justify-start">
				<RiSearchLine />
				<span class="text-muted-foreground">Find an entity…</span>
			</Button.Root>
		{/snippet}
	</Popover.Trigger>
	<Popover.Content align="start" class="w-96 p-0">
		<Command.Root shouldFilter={false}>
			<Command.Input bind:value={view.search} placeholder="Search names, kinds, or aliases…" />
			<Command.List class="max-h-80">
				{#if view.search.trim().length < 2}
					<Command.Empty>Enter at least two characters.</Command.Empty>
				{:else if view.searching}
					<Command.Loading>Searching…</Command.Loading>
				{:else if view.searchResults.length === 0}
					<Command.Empty>No entities found.</Command.Empty>
				{:else}
					<Command.Group heading="Knowledge entities">
						{#each view.searchResults as entity (entity.id)}
							<Command.Item value={entity.id} onclick={() => view.explore(entity)}>
								<div class="min-w-0">
									<div class="truncate text-sm font-medium">
										{entity.attributes.latestState?.displayName ||
											entity.attributes.aliases[0]?.attributes.providerSubjectRef ||
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
