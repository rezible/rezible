<script lang="ts">
	import Header from "$components/layout/header/Header.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import Stepper from "$components/layout/stepper/Stepper.svelte";

	import { initInitialSetupViewController } from "./controller.svelte";
	import IntegrationInstallDialog from "$features/settings/components/configure-integration-dialog/ConfigureIntegrationDialog.svelte";
	
	const ctrl = initInitialSetupViewController();
</script>

<IntegrationInstallDialog />


<div class="grid h-full w-full place-items-center">
	<div class="flex w-full max-w-4xl flex-col gap-4 border border-border bg-background p-4">
		<Header title="Initial Setup" classes={{ root: "gap-2", title: "text-2xl" }} />

		{#if ctrl.loading}
			<Spinner />
		{:else}
			<Stepper controller={ctrl.stepper} />
		{/if}
	</div>
</div>
