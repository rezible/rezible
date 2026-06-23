<script lang="ts">
	import * as Dialog from "$components/ui/dialog";
	import { Button } from "$components/ui/button";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";

	import IntegrationInstallTargetSelect from "./IntegrationInstallTargetSelect.svelte";
	import { initConfigureIntegrationDialogController } from "./controller.svelte";

	const ctrl = initConfigureIntegrationDialogController();

	const integration = $derived(ctrl.integration);

	const toTitleCase = (s: string) => s.toLowerCase().replace(/\b\w/g, (s) => s.toUpperCase());

	const title = $derived.by(() => {
		if (!integration) return;
		const name = toTitleCase(integration.displayName);
		const provider = toTitleCase(integration.provider);
		if (ctrl.installPending) return `Select Option to Install ${name}`;
		if (ctrl.oauthPending) return `Continuing with ${provider}`;
		if (ctrl.installation) return `Configure ${name} Installation`;
		return `Install ${name} Integration`;
	});
</script>

<Dialog.Root bind:open={() => ctrl.isOpen, (o) => ctrl.setOpen(o)}>
	<Dialog.Content class="max-h-[min(720px,calc(100vh-2rem))] overflow-y-auto sm:max-w-2xl">
		{#if !!integration}
			<Dialog.Header>
				<div class="flex flex-col gap-2 pr-8">
					<div class="flex flex-wrap items-center gap-2">
						<Dialog.Title>{title}</Dialog.Title>
					</div>
				</div>
			</Dialog.Header>

			<div class="flex flex-col gap-4">
				{#if ctrl.installationTargetSelectionRequired}
					<IntegrationInstallTargetSelect />
				{:else if ctrl.oauthError}
					<InlineAlert
						error={ctrl.oauthError}
						onDismiss={() => ctrl.integrations.oauth.clearFlow()}
					/>
				{:else if ctrl.oauthPending}
					<Spinner />
				{:else}
					{#if ctrl.configError}
						<InlineAlert error={ctrl.configError} />
					{/if}
					<ctrl.ConfigComponent />
				{/if}
			</div>
		{/if}

		{#if ctrl.isOpen && (ctrl.isEditMode || !ctrl.integration?.oauthInstall) && !ctrl.oauthPending}
			<Dialog.Footer>
				<div class="flex flex-wrap gap-2 items-center">
					<Button
						variant="outline"
						disabled={ctrl.loading}
						onclick={() => {
							ctrl.setOpen(false);
						}}
					>
						Cancel
					</Button>
					<Button disabled={!ctrl.hasChanged || !ctrl.configValid || ctrl.loading} onclick={() => ctrl.saveConfig()}>
						{#if ctrl.loading}
							<Spinner />
							{ctrl.installation ? "Sav" : "Install"}ing...
						{:else}
							{ctrl.installation ? "Save" : "Install"}
						{/if}
					</Button>
				</div>
			</Dialog.Footer>
		{/if}
	</Dialog.Content>
</Dialog.Root>
