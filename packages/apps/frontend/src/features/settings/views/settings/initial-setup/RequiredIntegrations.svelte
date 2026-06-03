<script lang="ts">
	import { useInitialSetupViewController } from "./controller.svelte";
	import AvailableIntegrationCard from "$src/features/settings/components/integration-provider-card/AvailableIntegrationCard.svelte";

	const ctrl = useInitialSetupViewController();
</script>

<div class="flex gap-2">
    <div class="flex flex-col gap-2">
        {#each ctrl.remainingRequiredCapabilities as capability}
            {@const available = ctrl.availableIntegrationsForCapabilities.get(capability)}
            {#each available as integration}
                {#key integration.name}
                    <AvailableIntegrationCard {integration} />
                {/key}
            {:else}
                <div class="p-2 border-error-300 border-2">
                    <span>No integrations available for this capability</span>
                </div>
            {/each}
        {/each}
    </div>
</div>