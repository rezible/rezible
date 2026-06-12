<script lang="ts">
	import { useInitialSetupViewController } from "../controller.svelte";
	import IntegrationProviderCard from "$features/settings/components/integration-provider-card/IntegrationProviderCard.svelte";

	const ctrl = useInitialSetupViewController();

	const nextRequired = $derived(ctrl.remainingRequiredCapabilities.at(0) || "");
	const availableForRequired = $derived(ctrl.availableIntegrationsForCapabilities.get(nextRequired) || []);
</script>

<div class="flex flex-col gap-4">
	<div class="space-y-1">
		<h2 class="text-lg font-semibold">Recommended integrations</h2>
		{#if nextRequired}
			<p class="text-sm text-muted-foreground">
				Configure an integration that provides {nextRequired} data, or skip this for now and add it later in
				Settings > Integrations.
			</p>
		{:else}
			<p class="text-sm text-muted-foreground">Required integration capabilities are configured.</p>
		{/if}
	</div>

	{#if nextRequired}
		<div class="flex flex-col gap-3">
			{#each availableForRequired as integration (integration.name)}
				<IntegrationProviderCard provider={integration.provider} onlyName={integration.name} showInstalledDetails={false} />
			{:else}
				<div class="border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
					No integrations are available for {nextRequired}. You can finish setup and add integrations later.
				</div>
			{/each}
		</div>
	{/if}
</div>
