<script lang="ts">
	import { Badge } from "$components/ui/badge";

    import type { InstalledIntegration } from "$lib/api";
	import Button from "$components/ui/button/button.svelte";
	import InstalledIntegrationDataSync from "./InstalledIntegrationDataSync.svelte";

	type Props = {
		installation: InstalledIntegration;
        openConfigDialog: () => void;
	};
	const { installation, openConfigDialog }: Props = $props();

    const attrs = $derived(installation.attributes);
    const enabledCapabilities = $derived(Object.entries(attrs.capabilities).filter(([, enabled]) => enabled));
</script>

<div class="flex flex-col gap-2 p-3 rounded-md border">
    <div class="flex items-center justify-between text-sm">
        <div class="min-w-0">
            <div class="truncate font-medium">{attrs.displayName}</div>
            <div class="truncate text-muted-foreground">{attrs.externalRef}</div>
        </div>
        <div class="flex flex-wrap justify-end gap-1">
            {#each enabledCapabilities as [cap] (cap)}
                <Badge variant="secondary">{cap}</Badge>
            {/each}
        </div>
    </div>
    <Button onclick={openConfigDialog}>Edit</Button>
    <InstalledIntegrationDataSync {installation} />
</div>