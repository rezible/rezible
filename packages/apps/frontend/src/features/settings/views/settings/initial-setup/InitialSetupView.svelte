<script lang="ts">
	import Header from "$components/layout/header/Header.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import Stepper from "$components/layout/stepper/Stepper.svelte";
	import { StepperController } from "$components/layout/stepper/stepper.svelte";

	import { initInitialSetupViewController } from "./controller.svelte";
	import OrganizationDetails from "./steps/OrganizationDetails.svelte";
	import WorkspacePreferences from "./steps/WorkspacePreferences.svelte";
	import RequiredCapabilities from "./steps/RequiredCapabilities.svelte";

	const ctrl = initInitialSetupViewController();

	const stepper = new StepperController({
		steps: [
			{
				label: "Details",
				description: "Organization and project details",
				component: OrganizationDetails,
				onNext: () => ctrl.onOrgDetailsNext(),
				canContinue: () => ctrl.canContinueOrgDetails,
			},
			{
				label: "Preferences",
				description: "Choose workspace behaviour",
				component: WorkspacePreferences,
				onNext: () => ctrl.onWorkspacePreferencesNext(),
				canContinue: () => ctrl.canContinueWorkspacePrefs,
			},
			{
				label: "Integrations",
				description: "Configure recommended integrations",
				component: RequiredCapabilities,
				canContinue: () => ctrl.canContinueCapabilities,
			},
		],
		onFinish: () => ctrl.doFinishOrganizationSetup(),
	});
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex w-full max-w-4xl flex-col gap-4 border border-border bg-background p-4">
		<Header title="Initial Setup" classes={{ root: "gap-2", title: "text-2xl" }} />

		{#if ctrl.loading}
			<Spinner />
		{:else}
			<Stepper controller={stepper} />
		{/if}
	</div>
</div>
