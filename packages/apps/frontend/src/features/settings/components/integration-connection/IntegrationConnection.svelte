<script lang="ts">
	import type { Snippet } from "svelte";
	import RiLinkUnlink from "remixicon-svelte/icons/link-unlink";
	import RiRefreshLine from "remixicon-svelte/icons/refresh-line";

	import type { IntegrationInstallation } from "$lib/api";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";

	import { useIntegrationDataSyncController } from "../integration-datasync-dialog/controller.svelte";
	import { useIntegrationProviderController } from "../integration-provider/controller.svelte";

	type Props = {
		installation: IntegrationInstallation;
		referenceLabel?: string;
		// Extra actions, such as opening settings or replacing credentials.
		actions?: Snippet;
		children?: Snippet;
	};
	const { installation, referenceLabel, actions, children }: Props = $props();

	const ctrl = useIntegrationProviderController();
	const sync = useIntegrationDataSyncController();

	const attributes = $derived(installation.attributes);
	const canSync = $derived(attributes.capabilities.includes("event_sync"));
</script>

<div class="flex flex-col gap-4 border p-4">
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div class="flex min-w-0 flex-col gap-1">
			<div class="flex min-w-0 items-center gap-2">
				<span class="truncate font-medium">{attributes.displayName}</span>
				<!-- Connection health from data sync will appear beside this status. -->
				<Badge variant="secondary">Installed</Badge>
			</div>
			{#if referenceLabel}
				<span class="truncate text-xs text-muted-foreground">
					{referenceLabel}
					{attributes.providerInstallationRef}
				</span>
			{/if}
		</div>

		<div class="flex flex-wrap gap-2">
			{@render actions?.()}
			{#if canSync}
				<Button variant="outline" size="sm" onclick={() => sync.openFor(installation)}>
					<RiRefreshLine />
					Sync
				</Button>
			{/if}
			<Button variant="outline" size="sm" onclick={() => ctrl.requestRemoval(installation)}>
				<RiLinkUnlink />
				Disconnect
			</Button>
		</div>
	</div>

	{@render children?.()}
</div>
