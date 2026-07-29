<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import Spinner from "$src/components/ui/spinner/spinner.svelte";
	import { initIntegrationProviderConfigController } from "./controller.svelte";
	import IntegrationDataSyncDialog from "$features/settings/components/integration-datasync-dialog/IntegrationDataSyncDialog.svelte";
	import { initIntegrationDataSyncController } from "$features/settings/components/integration-datasync-dialog/controller.svelte";
	import IntegrationInstallTargetSelect from "$features/settings/components/integration-install-target-selection/IntegrationInstallTargetSelect.svelte";

	type Props = {
		name: string;
	};
	const { name }: Props = $props();

	const controller = initIntegrationProviderConfigController(() => name);
	initIntegrationDataSyncController();
</script>

<div class="flex max-w-4xl flex-col gap-4">
	{#if controller.nameValid}
		{#if controller.loading}
			<Spinner />
		{:else if controller.ProviderComponent}
			{#each controller.availableIntegrations as integration (integration.name)}
				{@const targets = controller.installTargetOptionsFor(integration.name)}
				{#if targets.length > 0}
					<IntegrationInstallTargetSelect
						options={targets}
						onConfirm={(refs) =>
							controller.integrations.installFromTargets(integration.name, refs)}
					/>
				{/if}
			{/each}
			<controller.ProviderComponent />
		{:else}
			<InlineAlert
				error={{
					title: "Integration provider not found",
					detail: `No integration provider named "${name}" is available.`,
					status: 404,
				}}
			/>
		{/if}
	{/if}
</div>

<IntegrationDataSyncDialog />
