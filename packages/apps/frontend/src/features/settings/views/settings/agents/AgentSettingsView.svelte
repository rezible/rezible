<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";
	import { Badge } from "$components/ui/badge";
	import * as Card from "$components/ui/card";
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initAgentSettingsController } from "./controller.svelte";

	const view = initAgentSettingsController();

	registerPageDescriptor(() => ({
		title: "Agents",
		parents: [{ label: "Settings", path: resolve("/settings") }],
	}));
</script>

{#if view.session.isAdmin}
	<div class="flex max-w-4xl flex-col gap-4">
		{#if view.loading}
			<div class="flex items-center gap-2 text-sm text-muted-foreground">
				<LoadingIndicator />
				<span>Loading agents...</span>
			</div>
		{:else if view.error}
			<InlineAlert error={view.error} />
		{:else if view.agents}
			<Card.Root>
				<Card.Header>
					<Card.Title>Agents</Card.Title>
				</Card.Header>
				<Card.Content class="divide-y p-0">
					{#each view.agents as agent (agent.name)}
						<div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-4 px-6 py-3">
							<div class="min-w-0">
								<div class="truncate font-medium">{agent.displayName}</div>
								<div class="truncate text-sm text-muted-foreground">
									{"Ready"}
								</div>
							</div>
							<Badge>
								{"Ready"}
							</Badge>
						</div>
					{/each}
				</Card.Content>
			</Card.Root>
		{/if}
	</div>
{/if}
