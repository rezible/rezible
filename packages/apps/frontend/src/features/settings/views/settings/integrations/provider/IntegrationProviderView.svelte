<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";

	import { initIntegrationProviderController } from "$features/settings/components/integration-provider/controller.svelte";
	import IntegrationDataSyncDialog from "$features/settings/components/integration-datasync-dialog/IntegrationDataSyncDialog.svelte";
	import { initIntegrationDataSyncController } from "$features/settings/components/integration-datasync-dialog/controller.svelte";
	import IntegrationRemoveDialog from "$features/settings/components/integration-remove-dialog/IntegrationRemoveDialog.svelte";

	type Props = {
		name: string;
	};
	const { name }: Props = $props();

	const ctrl = initIntegrationProviderController(() => name);
	initIntegrationDataSyncController();

	registerPageDescriptor(() => ({
		title: ctrl.provider?.displayName ?? "Integration",
		parents: [
			{ label: "Settings", path: resolve("/settings") },
			{ label: "Organization", path: resolve("/settings/organization") },
			{ label: "Integrations", path: resolve("/settings/integrations") },
		],
	}));
</script>

<div class="flex max-w-3xl flex-col gap-6">
	{#if ctrl.loading}
		<div class="flex items-center gap-4">
			<Skeleton class="size-12" />
			<div class="flex flex-1 flex-col gap-2">
				<Skeleton class="h-5 w-40" />
				<Skeleton class="h-4 w-72 max-w-full" />
			</div>
		</div>
		<Skeleton class="h-48 w-full" />
	{:else if ctrl.error}
		<InlineAlert error={ctrl.error} dismissable={false} />
	{:else if !ctrl.provider || !ctrl.ProviderComponent}
		<InlineAlert
			error={{
				title: "Integration not found",
				detail: `No integration provider named "${name}" is available.`,
			}}
			dismissable={false}
		/>
		<Button class="w-fit" variant="outline" href={resolve("/settings/integrations")}>
			Back to integrations
		</Button>
	{:else}
		{@const provider = ctrl.provider}
		<header class="flex items-start gap-4">
			<div class="grid size-12 shrink-0 place-items-center border bg-muted">
				<provider.icon class="size-6" />
			</div>
			<div class="flex min-w-0 flex-col gap-1">
				<h1 class="text-lg font-semibold">{provider.displayName}</h1>
				<p class="text-sm text-muted-foreground">{provider.description}</p>
			</div>
		</header>

		<ctrl.ProviderComponent />
	{/if}
</div>

<IntegrationDataSyncDialog />
<IntegrationRemoveDialog />
