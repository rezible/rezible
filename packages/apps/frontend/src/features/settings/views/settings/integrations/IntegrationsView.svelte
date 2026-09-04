<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";

	import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

	import * as Card from "$components/ui/card";

	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { Badge } from "$src/components/ui/badge";
	import { Button } from "$src/components/ui/button";
	import { resolve } from "$app/paths";

	const controller = useIntegrationsController();

	registerPageDescriptor(() => ({
		title: "Integrations",
		parents: [{ label: "Settings", path: resolve("/settings") }],
	}));
</script>

<div class="flex flex-col gap-4">
	{#if controller.loading}
		<div class="flex items-center gap-2">
			<LoadingIndicator />
			<span>Loading...</span>
		</div>
	{:else if controller.error}
		<InlineAlert error={controller.error} />
	{:else}
		<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
			{#each controller.providers as provider (provider.name)}
				{@const hasInstalls = controller.installationsByProvider.has(provider.name)}
				<Card.Root>
					<Card.Header>
						<Card.Title class="truncate">{provider.displayName}</Card.Title>
						<Card.Action>
							<Badge variant={hasInstalls ? "default" : "outline"}>
								{hasInstalls ? "Installed" : "Not installed"}
							</Badge>
						</Card.Action>
					</Card.Header>
					<Card.Footer>
						<Button
							href={resolve("/settings/integrations/[provider]", { provider: provider.name })}
							variant="outline"
						>
							Configure
						</Button>
					</Card.Footer>
				</Card.Root>
			{/each}
		</div>
	{/if}
</div>
