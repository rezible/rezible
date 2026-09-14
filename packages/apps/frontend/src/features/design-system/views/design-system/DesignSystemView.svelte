<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	import FoundationsSection from "./sections/FoundationsSection.svelte";
	import ControlsSection from "./sections/ControlsSection.svelte";
	import ContentSection from "./sections/ContentSection.svelte";
	import OverlaysSection from "./sections/OverlaysSection.svelte";
	import DesignSystemActions from "./DesignSystemActions.svelte";

	registerPageDescriptor(() => ({
		title: "Design system",
		actions: { component: DesignSystemActions },
	}));
</script>

<svelte:head><title>Atlas design system · Rezible</title></svelte:head>

<FeatureNavigationRail
	route="/design-system/[[section=designSystemSection]]"
	label="Design system"
	entries={[
		{ label: "Foundations", component: FoundationsSection, params: { section: undefined } },
		{ label: "Controls", component: ControlsSection, params: { section: "controls" } },
		{ label: "Content", component: ContentSection, params: { section: "content" } },
		{ label: "Overlays", component: OverlaysSection, params: { section: "overlays" } },
	]}
></FeatureNavigationRail>
