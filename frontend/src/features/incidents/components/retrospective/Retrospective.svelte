<script lang="ts">
	import { onMount } from "svelte";
	import { watch } from "runed";

	import IncidentAnalysis from "./incident-analysis/IncidentAnalysis.svelte";
    import IncidentReport from "./incident-report/IncidentReport.svelte";
    import ContextSidebar from "./context-sidebar/ContextSidebar.svelte";

	import { collaborationState } from '$features/incidents/lib/collaboration.svelte';
	import { retrospectiveCtx } from '$features/incidents/lib/context.ts';

	type Props = {
		view: "analysis" | "report";
    }
    const { view }: Props = $props();

	const retrospectiveId = retrospectiveCtx.get().id;
	watch(() => retrospectiveId, id => {collaborationState.connect(id)});
	onMount(() => {return () => {collaborationState.cleanup()}});
</script>

<div class="flex-1 min-h-0 overflow-y-auto border p-2">
	{#if view === "analysis"}
		<IncidentAnalysis />
	{:else if view === "report"}
		<IncidentReport />
	{/if}
</div>

<ContextSidebar />
