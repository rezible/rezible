<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { useIntegrationCardController } from "./controller.svelte";

    import type { ConfiguredIntegration } from "$lib/api";

	type Props = {
		configured: ConfiguredIntegration;
	};
	const { configured: ci }: Props = $props();

	const ctrl = useIntegrationCardController();

    const attrs = $derived(ci.attributes);
</script>

<div class="flex items-center justify-between gap-3 rounded-md border p-3 text-sm">
    <div class="min-w-0">
        <div class="truncate font-medium">{attrs.displayName}</div>
        <div class="truncate text-muted-foreground">{attrs.externalRef}</div>
    </div>
    <div class="flex flex-wrap justify-end gap-1">
        {#each Object.entries(attrs.dataKinds).filter(([, enabled]) => enabled) as [kind] (kind)}
            <Badge variant="secondary">{kind}</Badge>
        {/each}
    </div>
</div>