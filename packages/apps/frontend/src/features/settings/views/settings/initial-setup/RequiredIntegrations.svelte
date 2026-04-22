<script lang="ts">
	import Header from "$components/header/Header.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import { initInitialSetupViewController } from "./controller.svelte";
	import IntegrationConfigCard from "$features/settings/components/integration-config-card/IntegrationConfigCard.svelte";

	const ctrl = initInitialSetupViewController();
</script>

<div class="flex gap-2">
    <div class="flex flex-col gap-2">
        {#each ctrl.availableDataKindIntegrations as integration}
            {@const name = integration.name}
            {#key name}
                <IntegrationConfigCard {integration} />
            {/key}
        {:else}
            <div class="p-2 border-error-300 border-2">
                <span>No integrations available for this data</span>
            </div>
        {/each}
    </div>
</div>