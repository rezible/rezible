<script lang="ts">
	import type { AvailableIntegration } from "$lib/api";

	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
	import InlineAlert from "$src/components/layout/error-alert/ErrorAlert.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";

	import { initAvailableIntegrationCardController } from "./availableIntegrationController.svelte";
	import InstalledIntegrationPanel from "./InstalledIntegrationPanel.svelte";
	import IntegrationInstallTargetSelect from "./IntegrationInstallTargetSelect.svelte";

	import RiCircleLine from "remixicon-svelte/icons/add-circle-line";

	type Props = {
		integration: AvailableIntegration;
	};
	const { integration }: Props = $props();

	const ctrl = initAvailableIntegrationCardController(() => integration);

	const showContent = $derived(!!ctrl.configError || ctrl.installationTargetSelectionRequired || ctrl.hasInstalled || ctrl.showConfig);
</script>

<Card.Root class="gap-4 p-4 min-w-xs">
    <Card.Header class="p-0">
        <div class="flex items-center justify-between gap-4 h-fit">
            <div class="flex flex-col gap-2">
                <Card.Title class="">{integration.name}</Card.Title>
            </div>
            <div class="flex items-center gap-2">
                <Button
                    onclick={() => {ctrl.startConfig()}}
                    variant="outline"
                    disabled={ctrl.loading}
                >
                    Connect {ctrl.hasInstalled ? "another" : ""}
                    <RiCircleLine />
                </Button>
            </div>
        </div>
    </Card.Header>

    <Card.Content class="p-0 flex flex-col gap-3 {showContent ? '' : 'hidden'}">
        {#if !!ctrl.configError}
            <InlineAlert bind:error={ctrl.configError} />
        {/if}

        {#if ctrl.installationTargetSelectionRequired}
            <IntegrationInstallTargetSelect />
        {:else if ctrl.showConfig}
            <ctrl.ConfigComponent />

            <div class="flex flex-wrap gap-2">
                <Button disabled={!ctrl.hasChanges || ctrl.loading} onclick={() => ctrl.saveConfig()}>
                    {#if ctrl.loading}
                        <Spinner />
                        Saving...
                    {:else}
                        Save
                    {/if}
                </Button>
                <Button disabled={ctrl.loading} onclick={() => ctrl.clearConfig()}>Cancel</Button>
            </div>
        {:else if ctrl.hasInstalled}
            <div class="flex flex-col gap-2">
                {#each ctrl.installations as installation (installation.id)}
                    <InstalledIntegrationPanel {installation} />
                {/each}
            </div>
        {:else}
            <span>No connections configured</span>
        {/if}
    </Card.Content>
</Card.Root>