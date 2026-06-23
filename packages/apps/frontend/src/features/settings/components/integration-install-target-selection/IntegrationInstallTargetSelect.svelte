<script lang="ts">
	import { Button } from "$components/ui/button";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import { Checkbox } from "$components/ui/checkbox";
	import type { IntegrationInstallTarget } from "@rezible/api-client-ts";
	import { SvelteSet } from "svelte/reactivity";
    
    type Props = {
        options: IntegrationInstallTarget[];
        onConfirm: (refs: string[]) => void;
    };
    const { options, onConfirm }: Props = $props();

    let selectedRefs = new SvelteSet<string>();

	const toggleInstallationTargetSelection = (ref: string, selected: boolean) => {
		if (selected) {
			selectedRefs.add(ref);
		} else {
			selectedRefs.delete(ref);
		}
	}
</script>


<div class="flex flex-col gap-3 rounded-md border p-3">
    <div class="flex flex-col gap-1">
        <span class="text-sm font-medium">Select installations</span>
        <span class="text-sm text-muted-foreground">Choose which accounts to connect.</span>
    </div>
    <div class="flex flex-col gap-2">
        {#each options as option (option.externalRef)}
            <label class="flex items-center gap-3 rounded-md border p-3 text-sm">
                <Checkbox
                    checked={selectedRefs.has(option.externalRef)}
                    onCheckedChange={(checked) =>
                        toggleInstallationTargetSelection(option.externalRef, !!checked)}
                />
                <span class="flex flex-col">
                    <span class="font-medium">{option.displayName}</span>
                    <span class="text-muted-foreground">{option.externalRef}</span>
                </span>
            </label>
        {/each}
    </div>
    <Button
        disabled={selectedRefs.size === 0}
        onclick={() => onConfirm(selectedRefs.values().toArray())}
    >
        Connect selected
    </Button>
</div>