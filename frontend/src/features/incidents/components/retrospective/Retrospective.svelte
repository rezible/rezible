<script lang="ts">
	import { onMount } from 'svelte';
	import { watch } from "runed";
    import { createQuery } from '@tanstack/svelte-query';

	import { getIncidentUserDebriefOptions, getRetrospectiveForIncidentOptions } from '$lib/api';
	import { onQueryUpdate } from '$lib/utils.svelte';
	import { collaborationState } from '$features/incidents/lib/collaboration.svelte';

	import EditorWrapper from './editor/EditorWrapper.svelte';
    import IncidentDebriefDialog from './debrief-dialog/IncidentDebriefDialog.svelte';

	type Props = { incidentId: string };
	let { incidentId }: Props = $props();

	const retroQuery = createQuery(() => getRetrospectiveForIncidentOptions({path: {id: incidentId}}));
	const retrospective = $derived(retroQuery.data?.data);
	const retrospectiveId = $derived(retrospective?.id);

	const documentName = $derived(retrospective?.attributes.documentName ?? "");
	watch(() => documentName, (name: string) => {collaborationState.connect(name)});
	// $effect(() => {if (documentName) collaborationState.connect(documentName)});
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

{#if sections}
	<EditorWrapper {sections} />
{/if}

{#if debrief}
	<IncidentDebriefDialog {debrief} bind:open={showDebriefDialog} />
{/if}