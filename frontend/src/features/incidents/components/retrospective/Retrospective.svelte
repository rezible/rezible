<script lang="ts">
	import { collaboration } from '$features/incidents/lib/collaboration.svelte';

	import IncidentAnalysis from "./incident-analysis/IncidentAnalysis.svelte";
    import IncidentReport from "./incident-report/IncidentReport.svelte";
    import ContextSidebar from "./context-sidebar/ContextSidebar.svelte";

	type Props = {
		view: "analysis" | "report";
    }
    const { view }: Props = $props();

	collaboration.setup();
</script>

<div class="flex-1 min-h-0 overflow-y-auto border p-2">
	{#if view === "analysis"}
		<IncidentAnalysis />
	{:else if view === "report"}
		<IncidentReport />
	{/if}
</div>

<ContextSidebar />
