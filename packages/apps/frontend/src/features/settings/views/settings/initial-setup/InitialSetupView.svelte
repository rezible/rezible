<script lang="ts">
	import Header from "$src/components/layout/header/Header.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import Stepper from "$components/layout/stepper/Stepper.svelte";
	import { StepperController } from "$components/layout/stepper/stepper.svelte";

	import { initInitialSetupViewController } from "./controller.svelte";
	import OrganizationDetails from "./steps/OrganizationDetails.svelte";
	import Preferences from "./steps/Preferences.svelte";
	import RequiredIntegrations from "./steps/RequiredCapabilities.svelte";

	const ctrl = initInitialSetupViewController();

	const stepper = new StepperController({
		steps: [
			{
				label: "Organization",
				description: "Name and workspace details",
				component: OrganizationDetails,
				canContinue: () => ctrl.canContinueOrgDetails,
			},
			{
				label: "Preferences",
				description: "Default workspace settings",
				component: Preferences,
				canContinue: () => ctrl.canContinuePreferences,
			},
			{
				label: "Integrations",
				description: "Recommended data sources",
				component: RequiredIntegrations,
				canContinue: () => ctrl.canFinish,
			},
		],
		onFinish: () => ctrl.doFinishOrganizationSetup(),
	});
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex w-full max-w-4xl flex-col gap-4 border border-border bg-background p-4">
		<Header title="Quick Setup" classes={{ root: "gap-2", title: "text-2xl" }} />

		{#if ctrl.loading}
			<Spinner />
		{:else}
			<Stepper controller={stepper} />
		{/if}
	</div>
</div>
