<script lang="ts">
	import { useInitialSetupViewController } from "./controller.svelte";
	import IntegrationCard from "$features/settings/components/integration-card/IntegrationCard.svelte";

	const ctrl = useInitialSetupViewController();
</script>

<div class="flex gap-2">
    <div class="flex flex-col gap-2">
        {#each ctrl.availableDataKindIntegrations as integration}
            {@const name = integration.name}
            {#key name}
                <IntegrationCard {integration} />
            {/key}
        {:else}
            <div class="p-2 border-error-300 border-2">
                <span>No integrations available for this data</span>
            </div>
        {/each}
    </div>
</div>