<script lang="ts">
	import { resolve } from "$app/paths";
	import RiFileTextLine from "remixicon-svelte/icons/file-text-line";
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { situationStatus } from "$features/situations/lib/status";

	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	import { initSituationController } from "./controller.svelte";
	import SituationOverviewView from "./overview/SituationOverviewView.svelte";
	import SituationInvestigationView from "./investigation/SituationInvestigationView.svelte";
	import SituationPageActions from "./SituationPageActions.svelte";
	import SituationActionDialogs from "./SituationActionDialogs.svelte";

	type Props = { id: string };
	const { id }: Props = $props();

	const controller = initSituationController(() => id);

	registerPageDescriptor(() => ({
		title: controller.situation?.attributes.title || "Situation",
		status: controller.situation ? situationStatus(controller.situation.attributes) : undefined,
		parents: [{ label: "Situations", path: resolve("/situations") }],
		related: controller.relatedLinks,
		contentHeading: true,
		pageActions: actions,
	}));
</script>

{#snippet actions()}
	<SituationPageActions {controller} />
{/snippet}

<FeatureNavigationRail
	label="Situation views"
	route="/situations/[id]/[[view=situationView]]"
	entries={[
		{
			label: "Brief",
			params: { id },
			icon: RiFileTextLine,
			component: SituationOverviewView,
		},
		{
			label: "Investigation",
			params: { id, view: "investigation" },
			icon: RiSearchLine,
			component: SituationInvestigationView,
		},
	]}
/>

<SituationActionDialogs {controller} />
