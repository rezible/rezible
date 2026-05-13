<script lang="ts">
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { initSettingsViewController } from "./controller.svelte";

	import General from "./general/General.svelte";
	import Integrations from "./integrations/Integrations.svelte";
	import InitialSetup from "./initial-setup/InitialSetup.svelte";
	import { initIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";

    const controller = initSettingsViewController();
	initIntegrationOAuthController();
</script>

{#if controller.showInitialSetup}
	<InitialSetup />
{:else}
	<TabbedViewContainer 
		route="/settings/[[view=settingsView]]"
		tabs={[
			{ label: "General", component: General, params: { view: undefined } },
			{ label: "Integrations", component: Integrations, params: { view: "integrations" } },
		]}
	/>
{/if}