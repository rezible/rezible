<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import Spinner from "$src/components/ui/spinner/spinner.svelte";
	import { initIntegrationProviderConfigController } from "./controller.svelte";

	type Props = {
		name: string;
	};
	const { name }: Props = $props();

	const controller = initIntegrationProviderConfigController(() => name);
</script>

<div class="flex max-w-4xl flex-col gap-4">
	{#if controller.nameValid}
		{#if controller.loading}
			<Spinner />
		{:else if controller.ProviderComponent}
			<controller.ProviderComponent />
		{:else}
			<InlineAlert error={{
				title: "Integration provider not found",
				detail: `No integration provider named "${name}" is available.`,
				status: 404,
			}} />
		{/if}
	{/if}
</div>
