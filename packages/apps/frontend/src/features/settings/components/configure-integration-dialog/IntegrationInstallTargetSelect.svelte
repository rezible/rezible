<script lang="ts">
	import { Button } from "$components/ui/button";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import { Checkbox } from "$components/ui/checkbox";
    
	import { useConfigureIntegrationDialogController } from "./controller.svelte";

    const ctrl = useConfigureIntegrationDialogController();
</script>


<div class="flex flex-col gap-3 rounded-md border p-3">
    <div class="flex flex-col gap-1">
        <span class="text-sm font-medium">Select installations</span>
        <span class="text-sm text-muted-foreground">Choose which accounts to connect.</span>
    </div>
    <div class="flex flex-col gap-2">
        {#each ctrl.installTargetOptions as option (option.externalRef)}
            <label class="flex items-center gap-3 rounded-md border p-3 text-sm">
                <Checkbox
                    checked={ctrl.selectedInstallTargetExternalRefs.has(option.externalRef)}
                    onCheckedChange={(checked) =>
                        ctrl.toggleInstallationTargetSelection(option.externalRef, !!checked)}
                />
                <span class="flex flex-col">
                    <span class="font-medium">{option.displayName}</span>
                    <span class="text-muted-foreground">{option.externalRef}</span>
                </span>
            </label>
        {/each}
    </div>
    <Button
        disabled={ctrl.loading || ctrl.selectedInstallTargetExternalRefs.size === 0}
        onclick={() => ctrl.confirmSelectedInstallationTargets()}
    >
        {#if ctrl.loading}
            <Spinner />
        {/if}
        Connect selected
    </Button>
</div>