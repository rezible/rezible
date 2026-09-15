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
	import SituationBrief from "./brief";
	import SituationImpact from "./impact";
	import SituationInvestigations from "./investigations";

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

<LoadingQueryWrapper query={controller.query} feedbackOnly />

<FeatureNavigationRail
	label="Situation views"
	route="/situations/[id]/[[view=situationView]]"
	entries={[
		{ label: "Brief", params: { id }, icon: RiFileTextLine, component: SituationBrief },
		{
			label: "Impact Scope",
			params: { id, view: "impact" },
			icon: RiFocus3Line,
			component: SituationImpact,
		},
		{
			label: "Investigations",
			params: { id, view: "investigations" },
			icon: RiSearchLine,
			component: SituationInvestigations,
		},
	]}
/>
