<script lang="ts">
	import { resolve } from "$app/paths";
	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";

	import RiFileTextLine from "remixicon-svelte/icons/file-text-line";
	import RiFocus3Line from "remixicon-svelte/icons/focus-3-line";
	import RiSearchLine from "remixicon-svelte/icons/search-line";

	import { initSituationController } from "./controller.svelte";

	import SituationPageActions from "./SituationPageActions.svelte";
	import SituationBriefView from "./brief/SituationBriefView.svelte";
	import SituationImpactView from "./impact/SituationImpactView.svelte";
	import SituationInvestigationView from "./investigation/SituationInvestigationView.svelte";

	type Props = { id: string };
	let { id }: Props = $props();

	const controller = initSituationController(() => id);

	registerPageDescriptor(() => ({
		title: controller.situation?.attributes.title || "Situation",
		parents: [{ label: "Situations", path: resolve("/situations") }],
		pageActions: actions,
	}));
</script>

{#snippet actions()}
	<SituationPageActions {controller} />
{/snippet}

<LoadingQueryWrapper query={controller.situationQuery} feedbackOnly />

<FeatureNavigationRail
	label="Situation views"
	route="/situations/[id]/[[view=situationView]]"
	entries={[
		{ 
			label: "Brief", 
			params: { id }, 
			icon: RiFileTextLine, 
			component: SituationBriefView,
		},
		{
			label: "Impact Scope",
			params: { id, view: "impact" },
			icon: RiFocus3Line,
			component: SituationImpactView,
		},
		{
			label: "Investigation",
			params: { id, view: "investigation" },
			icon: RiSearchLine,
			component: SituationInvestigationView,
		},
	]}
/>
