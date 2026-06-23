<script lang="ts">
	import * as Dialog from "$components/ui/dialog";
	import { Button } from "$components/ui/button";

	import IntegrationProviderConfig from "$features/settings/components/integration-provider/IntegrationProviderConfig.svelte";
	import { useInitialSetupController } from "../../initialSetupController.svelte";

    const ctrl = useInitialSetupController();

    const provider = $derived(ctrl.configureProvider);
    const setOpen = (o: boolean) => {
        if (!o) ctrl.configureProvider = undefined;
    }
</script>

<Dialog.Root bind:open={() => !!provider, setOpen}>
	<Dialog.Content class="max-h-[min(720px,calc(100vh-2rem))] overflow-y-auto sm:max-w-2xl">
        {#if !!provider}
            <Dialog.Header>
                <div class="flex flex-col gap-2 pr-8">
                    <div class="flex flex-wrap items-center gap-2">
                        <Dialog.Title>{provider.displayName}</Dialog.Title>
                    </div>
                </div>
            </Dialog.Header>

            <div class="flex flex-col gap-4">
                <IntegrationProviderConfig name={provider.name} />
            </div>

            <Dialog.Footer>
                <Button
                    onclick={() => {setOpen(false)}}
                    variant="outline"
                >
                    Close
                </Button>
            </Dialog.Footer>
        {/if}
	</Dialog.Content>
</Dialog.Root>
