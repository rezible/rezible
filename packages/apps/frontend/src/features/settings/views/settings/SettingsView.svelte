<script lang="ts">
	import TabbedViewContainer from "$src/components/layout/tabbed-view-container/TabbedViewContainer.svelte";
	import { initSettingsViewController } from "./controller.svelte";

	import General from "./general/General.svelte";
	import Integrations from "./integrations/Integrations.svelte";
	import InitialSetupView from "./initial-setup/InitialSetupView.svelte";

    const controller = initSettingsViewController();
</script>

{#if controller.showInitialSetup}
	<InitialSetupView />
{:else}
	<TabbedViewContainer 
		route="/settings/[[view=settingsView]]"
		tabs={[
			{ label: "General", component: General, params: { view: undefined } },
			{ label: "Integrations", component: Integrations, params: { view: "integrations" } },
		]}
	/>
{/if}