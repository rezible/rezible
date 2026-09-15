<script lang="ts">
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initInitialSetupController } from "./initialSetupController.svelte";
	import InitialSetupFormCard from "./InitialSetupFormCard.svelte";

	const ctrl = initInitialSetupController();

	registerPageDescriptor(() => ({ title: "Initial setup" }));
</script>

<div class="grid min-h-full w-full place-items-center p-4">
	{#if ctrl.loading}
		<Spinner aria-label="Loading setup" />
	{:else}
		<InitialSetupFormCard />
	{/if}
</div>
