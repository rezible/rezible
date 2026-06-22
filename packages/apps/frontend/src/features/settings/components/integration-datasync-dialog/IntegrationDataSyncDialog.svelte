<script lang="ts">
	import * as Dialog from "$components/ui/dialog";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";

	import { initIntegrationDataSyncController } from "./controller.svelte";

    const ctrl = initIntegrationDataSyncController();
</script>

<Dialog.Root bind:open={() => ctrl.isOpen, o => ctrl.setOpen(o)}>
	<Dialog.Content class="max-h-[min(720px,calc(100vh-2rem))] overflow-y-auto sm:max-w-2xl">
		{#if !!ctrl.installation}
			<Dialog.Header>
				<div class="flex flex-col gap-2 pr-8">
					<div class="flex flex-wrap items-center gap-2">
						<Dialog.Title>Data Sync - {ctrl.installation.attributes.displayName}</Dialog.Title>
					</div>
				</div>
			</Dialog.Header>

			<div class="flex flex-col gap-4">
                {#if !!ctrl.syncRunsError}
                    <InlineAlert error={ctrl.syncRunsError} dismissable={false} />
                {/if}

                {#if ctrl.latestSyncRunDisplay}
                    <Badge
                        variant={ctrl.latestSyncRunDisplay.variant}
                        class={ctrl.latestSyncRunDisplay.class}
                    >
                        {ctrl.latestSyncRunDisplay.label}
                    </Badge>
                {:else}
                    <Badge variant="outline">No runs</Badge>
                {/if}
                {#if ctrl.isLoading}
                    <Spinner aria-label="Sync status updating" />
                {/if}
			</div>

            <div class="flex min-w-0 items-center gap-2">
                <Button
                    onclick={() => {ctrl.requestSync()}}
                    variant="outline"
                    disabled={ctrl.disabled}
                >
                    {#if ctrl.isSyncing}
                        <Spinner />
                        Syncing...
                    {:else}
                        Request Data Sync
                    {/if}
                </Button>
            </div>
		{/if}
	</Dialog.Content>
</Dialog.Root>
