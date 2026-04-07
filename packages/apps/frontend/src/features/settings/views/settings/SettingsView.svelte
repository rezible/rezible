<script lang="ts">
	import type { SettingsViewParam } from "$src/params/settingsView";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { initSettingsViewController } from "./controller.svelte";

	import General from "./general/General.svelte";
	import Integrations from "./integrations/Integrations.svelte";
	import InitialSetup from "./initial-setup/InitialSetup.svelte";

    const controller = initSettingsViewController();

	const tabs: Tab<SettingsViewParam>[] = [
		{ label: "General", view: undefined, component: General },
		{ label: "Integrations", view: "integrations", component: Integrations },
	];
</script>

{#if controller.showInitialSetup}
    <InitialSetup />
{:else}
    <TabbedViewContainer path="/settings" {tabs} />
{/if}