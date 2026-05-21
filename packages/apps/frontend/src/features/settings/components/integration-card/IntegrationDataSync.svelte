<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { IntegrationDataSyncController, useIntegrationCardController } from "./controller.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";

	const ctrl = useIntegrationCardController();
    const syncCtrl = new IntegrationDataSyncController();
</script>

{#if !!syncCtrl.syncStatusError}
    <InlineAlert error={syncCtrl.syncStatusError} dismissable={false} />
{/if}

<div class="flex items-center justify-between gap-3 rounded-md border p-3 text-sm">
    <div class="flex min-w-0 items-center gap-2">
        <span class="text-muted-foreground">Data sync</span>
        {#if syncCtrl.latestSyncStatusDisplay}
            <Badge
                variant={syncCtrl.latestSyncStatusDisplay.variant}
                class={syncCtrl.latestSyncStatusDisplay.class}
            >
                {syncCtrl.latestSyncStatusDisplay.label}
            </Badge>
        {:else}
            <Badge variant="outline">No runs</Badge>
        {/if}
        {#if syncCtrl.isSyncing}
            <Spinner aria-label="Sync status updating" />
        {/if}
    </div>
    <div class="flex min-w-0 items-center gap-2">
        <Button
            onclick={() => {
                syncCtrl.requestSync();
            }}
            variant="outline"
            disabled={!ctrl.hasConfigured || ctrl.loading || syncCtrl.isSyncing}
        >
            {#if syncCtrl.isSyncing}
                <Spinner />
                Syncing...
            {:else}
                Request Data Sync
            {/if}
        </Button>
    </div>
</div>