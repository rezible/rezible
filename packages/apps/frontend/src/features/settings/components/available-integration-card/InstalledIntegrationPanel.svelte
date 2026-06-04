<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { useAvailableIntegrationCardController } from "./availableIntegrationController.svelte";

    import type { InstalledIntegration } from "$lib/api";
	import { initInstalledIntegrationCardController } from "./installedIntegrationController.svelte";
	import InstalledIntegrationDataSync from "./InstalledIntegrationDataSync.svelte";

	type Props = {
		installation: InstalledIntegration;
	};
	const { installation }: Props = $props();

    const ctrl = initInstalledIntegrationCardController(() => installation);

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
    <InstalledIntegrationDataSync {installation} />
</div>