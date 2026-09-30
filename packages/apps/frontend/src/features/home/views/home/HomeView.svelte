<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import PageCanvas from "$components/layout/page-canvas/PageCanvas.svelte";
	import { initHomeController } from "./controller.svelte";
	import NeedsYouSection from "./NeedsYouSection.svelte";
	import IncidentsSection from "./IncidentsSection.svelte";
	import SituationsSection from "./SituationsSection.svelte";
	import ActivitySection from "./ActivitySection.svelte";

	initHomeController();
	registerPageDescriptor(() => ({ title: "Home" }));
</script>

{#snippet context()}
	<ActivitySection />
{/snippet}

<PageCanvas {context} contextLabel="Recent activity">
	<NeedsYouSection />
	<IncidentsSection />
	<SituationsSection />
</PageCanvas>
