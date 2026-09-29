<script lang="ts">
	import { resolve } from "$app/paths";
	import RiFileTextLine from "remixicon-svelte/icons/file-text-line";
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import { registerPageDescriptor } from "$lib/app-shell.svelte"

	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import * as Empty from "$components/ui/empty";
	;
	import { initSituationController } from "./controller.svelte";
	import SituationPageActions from "./SituationPageActions.svelte";
	import SituationOverviewView from "./overview/SituationOverviewView.svelte";
	import SituationInvestigationView from "./investigation/SituationInvestigationView.svelte";

	type Props = { id: string };
	const { id }: Props = $props();

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

{#if controller.situationUnavailable}
	<Empty.Root>
		<Empty.Header>
			<Empty.Title>Situation unavailable</Empty.Title>
			<Empty.Description>
				This situation could not be found or is no longer accessible.
			</Empty.Description>
		</Empty.Header>
	</Empty.Root>
{:else}
	<LoadingQueryWrapper query={controller.situationQuery} feedbackOnly />
	<FeatureNavigationRail
		label="Situation views"
		route="/situations/[id]/[[view=situationView]]"
		entries={[
			{
				label: "Overview",
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
{/if}
