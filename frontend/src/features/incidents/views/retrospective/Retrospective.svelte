<script lang="ts">
	import './styles.postcss';
	import { onMount } from 'svelte';
    import { createQuery } from '@tanstack/svelte-query';
	import type { Editor as TiptapEditor } from '@tiptap/core';
	import { cls } from 'svelte-ux';

	import { getIncidentUserDebriefOptions, getRetrospectiveForIncidentOptions } from '$lib/api';

	import { activeEditor } from './editor.svelte';
	import { collaborationState } from './collaboration.svelte';

	import EditorWrapper from './EditorWrapper.svelte';
	import { DiscussionSidebar, draft } from './Discussions';
	import IncidentTimeline from './IncidentTimeline';
	import SectionsSidebar from './SectionsSidebar.svelte';
    import IncidentDebriefDialog from './IncidentDebrief';
    import type { AnnotationType } from './BubbleMenu.svelte';

	type Props = { incidentId: string };
	let { incidentId }: Props = $props();

	const retroQuery = createQuery(() => getRetrospectiveForIncidentOptions({path: {id: incidentId}}));
	const retrospective = $derived(retroQuery.data?.data);
	const documentName = $derived(retrospective?.attributes.documentName);
	$effect(() => {if (documentName) collaborationState.connect(documentName)});
	onMount(() => {return () => {collaborationState.cleanup()}});

	const sections = $derived(retrospective?.attributes.sections);

	const debriefQuery = createQuery(() => getIncidentUserDebriefOptions({path: {id: incidentId}}));
	const debrief = $derived(debriefQuery.data?.data);
	let showDebriefDialog = $state(false);

	let sectionsHidden = $state(false);

	let containerEl = $state<HTMLElement>();
	const sectionElements = $state<Record<string, HTMLElement>>({});

	const focusSectionFn = $state<Record<string, VoidFunction>>({});
	const onSectionClicked = (field: string) => {
		if (focusSectionFn[field]) focusSectionFn[field]();
		if (sectionElements[field]) sectionElements[field].scrollIntoView();
	}

	const onCreateAnnotation = (e: TiptapEditor, t: AnnotationType) => {
		if (t === 'draft-comment') {
			draft.create(e);
		} else if (t === 'service') {

		}
	}
</script>

<a href="http://localhost:5173/incidents/foo-bar-4/retrospective">other</a>

<div class="w-full flex-1 min-h-0 grid grid-cols-8 py-2 overflow-y-auto">
	<div class="col-span-1 grow block overflow-y-hidden max-h-full" class:hidden={sectionsHidden}>
		{#if sections && containerEl}
			<SectionsSidebar 
				bind:hidden={sectionsHidden}
				{containerEl} {sections} {sectionElements} {onSectionClicked} />
		{/if}
	</div>

	<div class={cls(
		"flex flex-col grow min-h-0 overflow-y-auto bg-surface-300", 
		sectionsHidden ? "col-start-1 col-span-5 border-3" : "col-span-4"
	)}>
		<div class="w-full overflow-y-auto pb-2 px-4 flex flex-col gap-4" bind:this={containerEl}>
			{#if sections && collaborationState.provider}
				{#each sections as section, i}
					<div bind:this={sectionElements[section.field]}>
						{#if section.type == "timeline"}
							<IncidentTimeline {incidentId} />
						{:else if section.type === "field"}
							<EditorWrapper
								provider={collaborationState.provider}
								{section}
								setIsActive={activeEditor.set}
								{onCreateAnnotation}
								bind:focusEditor={focusSectionFn[section.field]} 
							/>
						{/if}
					</div>
				{/each}
			{:else}
				<span>connecting provider</span>
			{/if}
		</div>
	</div>

	<div class="col-span-3 flex flex-col grow min-h-0 overflow-y-auto bg-surface-200 shadow-lg p-3">
		{#if retrospective && debrief}
			<DiscussionSidebar 
				bind:showDebriefDialog
				retrospectiveId={retrospective.id} 
				debriefId={debrief.id} 
			/>
		{/if}
	</div>
</div>

{#if debrief}
	<IncidentDebriefDialog {debrief} bind:open={showDebriefDialog} />
{/if}