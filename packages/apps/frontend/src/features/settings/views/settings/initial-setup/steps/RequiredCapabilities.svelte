<script lang="ts">
	import { useInitialSetupViewController } from "../controller.svelte";
	import AvailableIntegrationCard from "$features/settings/components/available-integration-card/AvailableIntegrationCard.svelte";

	const ctrl = useInitialSetupViewController();

	const nextRequiredCapability = $derived(ctrl.remainingRequiredCapabilities.at(0));
	const availableForRequired = $derived(
		(nextRequiredCapability ? ctrl.availableIntegrationsForCapabilities.get(nextRequiredCapability) : []) || [],
	);
</script>

<div class="flex flex-col gap-4">
	<div class="space-y-1">
		<h2 class="text-lg font-semibold">Recommended integrations</h2>
		{#if nextRequiredCapability}
			<p class="text-sm text-muted-foreground">Configure an integration that provides {nextRequiredCapability} data.</p>
		{:else}
			<p class="text-sm text-muted-foreground">Required integration capabilities are configured.</p>
		{/if}
	</div>

	{#if nextRequiredCapability}
		<div class="flex flex-col gap-3">
			{#each availableForRequired as integration (integration.name)}
				<AvailableIntegrationCard {integration} />
			{:else}
				<div class="border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
					No integrations are available for {nextRequiredCapability}.
				</div>
			{/each}
		</div>
	{/if}
</div>
