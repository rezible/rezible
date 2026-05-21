<script lang="ts">
	import { Button } from "$components/ui/button";
	import { useIntegrationCardController } from "./controller.svelte";
	import Spinner from "$src/components/ui/spinner/spinner.svelte";
	import { Checkbox } from "$components/ui/checkbox";
	import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";

    const ctrl = useIntegrationCardController();
	const oauth = useIntegrationOAuthController();
</script>


<div class="flex flex-col gap-3 rounded-md border p-3">
    <div class="flex flex-col gap-1">
        <span class="text-sm font-medium">Select installations</span>
        <span class="text-sm text-muted-foreground">Choose which accounts to connect.</span>
    </div>
    <div class="flex flex-col gap-2">
        {#each oauth.selectionOptions as option (option.externalRef)}
            <label class="flex items-center gap-3 rounded-md border p-3 text-sm">
                <Checkbox
                    checked={oauth.selectedExternalRefs.has(option.externalRef)}
                    onCheckedChange={(checked) =>
                        oauth.toggleSelection(option.externalRef, !!checked)}
                />
                <span class="flex flex-col">
                    <span class="font-medium">{option.displayName}</span>
                    <span class="text-muted-foreground">{option.externalRef}</span>
                </span>
            </label>
        {/each}
    </div>
    <Button
        disabled={ctrl.loading || oauth.selectedExternalRefs.size === 0}
        onclick={() => oauth.selectOAuthOptions()}
    >
        {#if ctrl.loading}
            <Spinner />
        {/if}
        Connect selected
    </Button>
</div>