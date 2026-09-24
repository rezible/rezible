<script lang="ts">
	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import type { SystemViewParam } from "$params/systemView";
	import SystemMapView from "./map";
	import CatalogueScreen from "./catalogue/CatalogueScreen.svelte";
	import SavedViewsScreen from "./saved-views/SavedViewsScreen.svelte";
	import { initSystemViewController } from "./controller.svelte";

	type Props = { view: SystemViewParam };
	let { view }: Props = $props();

	const controller = initSystemViewController(() => view);
	registerPageDescriptor(() => ({ title: controller.title }));
</script>

<FeatureNavigationRail
	label="System views"
	route="/system/[[view=systemView]]"
	entries={[
		{ label: "Map", params: {}, component: SystemMapView },
		{ label: "Catalogue", params: { view: "catalogue" }, component: CatalogueScreen },
		{ label: "Saved Views", params: { view: "saved-views" }, component: SavedViewsScreen },
	]}
/>
