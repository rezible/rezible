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

	type Props = { ref: string };
	const { ref }: Props = $props();

	const controller = initSituationController(() => ref);

	registerPageDescriptor(() => ({
		title: controller.situation?.attributes.title || "Situation",
		reference: controller.reference,
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
	route="/situations/[ref]/[[view=situationView]]"
	entries={[
		{
			label: "Brief",
			params: { ref },
			icon: RiFileTextLine,
			component: SituationOverviewView,
		},
		{
			label: "Investigation",
			params: { ref, view: "investigation" },
			icon: RiSearchLine,
			component: SituationInvestigationView,
		},
	]}
/>

<SituationActionDialogs {controller} />
