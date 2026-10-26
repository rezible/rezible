<script lang="ts">
	import { resolve } from "$app/paths";
	import RiDashboardLine from "remixicon-svelte/icons/dashboard-horizontal-line";
	import RiFireLine from "remixicon-svelte/icons/fire-line";
	import RiHistoryLine from "remixicon-svelte/icons/history-line";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	import { initAlertViewController } from "./controller.svelte";
	import AlertOverview from "./overview/AlertOverview.svelte";
	import AlertEvents from "./events/AlertEvents.svelte";
	import AlertIncidents from "./incidents/AlertIncidents.svelte";

	type Props = {
		id: string;
	};
	const { id }: Props = $props();

	const view = initAlertViewController(() => id);

	registerPageDescriptor(() => ({
		title: view.alertTitle || "Signal",
		parents: [{ label: "Signals", path: resolve("/signals") }],
		contentHeading: view.activeView === undefined,
	}));
</script>

<LoadingQueryWrapper query={view.alertQuery} feedbackOnly />

{#if view.alert}
	<FeatureNavigationRail
		label="Signal views"
		route="/signals/[id]/[[view=signalView]]"
		entries={[
			{
				label: "Overview",
				params: { id },
				icon: RiDashboardLine,
				component: AlertOverview,
			},
			{
				label: "Recent activity",
				params: { id, view: "events" },
				icon: RiHistoryLine,
				component: AlertEvents,
			},
			{
				label: "Incidents",
				params: { id, view: "incidents" },
				icon: RiFireLine,
				component: AlertIncidents,
			},
		]}
	/>
{/if}
