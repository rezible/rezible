<script lang="ts">
	import { Badge } from "$components/ui/badge";

    import type { InstalledIntegration } from "$lib/api";
	import Button from "$components/ui/button/button.svelte";
	import { useIntegrationsController } from "../../lib/integrationsController.svelte";

	type Props = {
		installation: InstalledIntegration;
        openConfigDialog: () => void;
	};
	const { installation, openConfigDialog }: Props = $props();

    const intgs = useIntegrationsController();

    const attrs = $derived(installation.attributes);
</script>

<div class="flex flex-col gap-2 p-3 rounded-md border">
    <div class="flex items-center justify-between text-sm">
        <div class="min-w-0">
            <div class="truncate font-medium">{attrs.displayName}</div>
            <div class="truncate text-muted-foreground">{attrs.externalRef}</div>
        </div>
        <div class="flex flex-wrap justify-end gap-1">
            <Button variant="secondary" onclick={() => {intgs.dataSyncDialogInstallation = installation}}>Data Sync</Button>
            <Button onclick={openConfigDialog}>Configure</Button>
        </div>
    </div>
</div>