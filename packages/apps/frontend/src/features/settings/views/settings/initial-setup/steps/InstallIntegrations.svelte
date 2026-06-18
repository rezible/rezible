<script lang="ts">
	import IntegrationProviderCard from "$features/settings/components/integration-provider-card/IntegrationProviderCard.svelte";
	import { useIntegrationsController } from "$src/features/settings/lib/integrationsController.svelte";
	import { useInitialSetupController } from "../initialSetupController.svelte";

	const ctrl = useInitialSetupController();
	const integrations = useIntegrationsController();
</script>

{#if ctrl.integrationSuggestions.length > 0}
	<div class="flex flex-row gap-4 justify-between">
		<div class="space-y-1 max-w-xs">
			<h2 class="text-lg font-semibold">Suggested integrations</h2>
			<p class="text-sm text-muted-foreground">
				Configure commonly useful integrations, 
				or skip this for now and add it later in Settings > Integrations.
			</p>
		</div>
		
		<div class="flex flex-col gap-3 flex-1">
			{#each ctrl.integrationSuggestions as suggestion}
				<div class="flex flex-col gap-1 border p-2 pt-1">
					<span class="text-lg">{suggestion.label}</span>

					{#each suggestion.available as integration (integration.name)}
						<IntegrationProviderCard provider={integration.provider} onlyName={integration.name} compact />
					{:else}
						<div class="border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
							No {suggestion.label} integrations are available.
						</div>
					{/each}
				</div>
			{/each}
		</div>
	</div>
{:else}
	<div class="space-y-1">
		<h2 class="text-lg font-semibold">All suggested integrations installed!</h2>
		<p class="text-sm text-muted-foreground">
			More integrations can be configured later in Settings > Integrations.
		</p>
	</div>
{/if}