<script lang="ts">
    import type { InstalledIntegration } from "$lib/api";

	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import InlineAlert from "$src/components/layout/error-alert/ErrorAlert.svelte";

	import { InstalledIntegrationDataSyncController } from "./availableIntegrationController.svelte";

	type Props = {
		installation: InstalledIntegration;
	};
	const { installation }: Props = $props();
    
    const ctrl = new InstalledIntegrationDataSyncController(() => installation);
</script>

{#if !!ctrl.syncStatusError}
    <InlineAlert error={ctrl.syncStatusError} dismissable={false} />
{/if}

<div class="flex items-center justify-between gap-3 rounded-md border p-3 text-sm">
    <div class="flex min-w-0 items-center gap-2">
        <span class="text-muted-foreground">Data sync</span>
        {#if ctrl.latestSyncStatusDisplay}
            <Badge
                variant={ctrl.latestSyncStatusDisplay.variant}
                class={ctrl.latestSyncStatusDisplay.class}
            >
                {ctrl.latestSyncStatusDisplay.label}
            </Badge>
        {:else}
            <Badge variant="outline">No runs</Badge>
        {/if}
        {#if ctrl.isSyncing}
            <Spinner aria-label="Sync status updating" />
        {/if}
    </div>
    <div class="flex min-w-0 items-center gap-2">
        <Button
            onclick={() => {
                ctrl.requestSync();
            }}
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
</div>