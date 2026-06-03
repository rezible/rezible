<script lang="ts">
	import { useInitialSetupViewController } from "./controller.svelte";
	import AvailableIntegrationCard from "$src/features/settings/components/integration-provider-card/AvailableIntegrationCard.svelte";

	const ctrl = useInitialSetupViewController();

    const nextRequiredCapability = $derived(ctrl.remainingRequiredCapabilities.at(0));
    const availableForRequired = $derived((!!nextRequiredCapability ? ctrl.availableIntegrationsForCapabilities.get(nextRequiredCapability) : []) || []);
    // const availableOptions = $derived(availableForRequired.length > 0 ? availableForRequired : ctrl.availableOptions);
</script>

<div class="flex gap-2">
    <div class="flex flex-col gap-2">
        <span>Capability: {nextRequiredCapability}</span>
        {#each availableForRequired as integration}
            {#key integration.name}
                <AvailableIntegrationCard {integration} />
            {/key}
        {:else}
            <div class="p-2 border-error-300 border-2">
                <span>No integrations available for this capability</span>
            </div>
        {/each}
    </div>
</div>