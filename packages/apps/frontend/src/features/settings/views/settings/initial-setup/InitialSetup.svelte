<script lang="ts">
	import Header from "$components/header/Header.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import Button from "$components/ui/button/button.svelte";
	import { initInitialSetupViewController } from "./controller.svelte";
	import RequiredIntegrations from "./RequiredIntegrations.svelte";

	const ctrl = initInitialSetupViewController();
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Quick Setup" classes={{ root: "gap-2", title: "text-2xl" }} />

		{#if ctrl.loading}
			<Spinner />
		{:else}
			{#if ctrl.step === "required_integrations"}
				<RequiredIntegrations />
			{:else}
				<Button href="/">Unknown state - go home</Button>
			{/if}
		{/if}
	</div>
</div>