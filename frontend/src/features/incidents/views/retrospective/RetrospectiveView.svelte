<script lang="ts">
	import { onMount } from 'svelte';
    import { createQuery } from '@tanstack/svelte-query';

	import { getIncidentUserDebriefOptions, getRetrospectiveForIncidentOptions } from '$lib/api';
	import { onQueryUpdate } from '$lib/utils.svelte';
	import { collaborationState } from './lib/collaboration.svelte';

	import EditorWrapper from './components/editor/EditorWrapper.svelte';
    import IncidentDebriefDialog from './components/debrief-dialog/IncidentDebriefDialog.svelte';
    import RetrospectiveSidebar from './components/sidebar/RetrospectiveSidebar.svelte';

	type Props = { incidentId: string };
	let { incidentId }: Props = $props();

	const retroQuery = createQuery(() => getRetrospectiveForIncidentOptions({path: {id: incidentId}}));
	const retrospective = $derived(retroQuery.data?.data);
	const retrospectiveId = $derived(retrospective?.id);

	const documentName = $derived(retrospective?.attributes.documentName);
	$effect(() => {if (documentName) collaborationState.connect(documentName)});
	onMount(() => {return () => {collaborationState.cleanup()}});

	const debriefQueryOpts = () => getIncidentUserDebriefOptions({path: {id: incidentId}});
	const debriefQuery = createQuery(debriefQueryOpts);
	const debrief = $derived(debriefQuery.data?.data);
	const debriefId = $derived(debrief?.id);

	let showDebriefDialog = $state(false);
	onQueryUpdate(debriefQueryOpts, ({data}) => {
		if (data.attributes.started || !data.attributes.required) return;
		showDebriefDialog = true;
	});

	const sections = $derived(retrospective?.attributes.sections);
</script>

<div class="w-full flex-1 min-h-0 grid grid-cols-9 py-2 overflow-y-auto">
	{#if sections}
		<EditorWrapper {sections} />
	{/if}

	{#if retrospectiveId && debriefId}
		<RetrospectiveSidebar 
			{incidentId}
			{retrospectiveId} 
			{debriefId} 
			bind:showDebriefDialog
		/>
	{/if}
</div>

{#if debrief}
	<IncidentDebriefDialog {debrief} bind:open={showDebriefDialog} />
{/if}