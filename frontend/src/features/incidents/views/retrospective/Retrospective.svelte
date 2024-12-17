<script lang="ts">
	import './styles.postcss';
	import { onMount } from 'svelte';

	import { Dialog, Header, Button, cls } from 'svelte-ux';
    import { mdiClose } from '@mdi/js';

	import type { Incident, IncidentDebrief, Retrospective } from '$lib/api';

	import type { Editor as TiptapEditor } from '@tiptap/core';
	import { activeEditor } from './editor.svelte';
	import { collaborationState } from './collaboration.svelte';

	import EditorWrapper from './EditorWrapper.svelte';
	import { DiscussionSidebar, draft } from './Discussions';
	import IncidentTimeline from './IncidentTimeline';
	import SectionsSidebar from './SectionsSidebar.svelte';
    import IncidentDebriefDialog from './IncidentDebrief';
    import type { AnnotationType } from './BubbleMenu.svelte';

	type Props = { 
		incident: Incident,
		retrospective: Retrospective,
		debrief: IncidentDebrief,
	};
	let { incident, retrospective, debrief }: Props = $props();

	const sections = $derived(retrospective.attributes.sections);

	let showDebrief = $state(!debrief.attributes.started);

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

	const documentName = $derived(retrospective.attributes.documentName);
	onMount(() => {
		collaborationState.connect(documentName);
		return () => {collaborationState.disconnect()};
	});

	let sectionsHidden = $state(false);
</script>

<div class="w-full flex-1 min-h-0 grid grid-cols-8 py-2 overflow-y-auto">
	<div class="col-span-1 grow block overflow-y-hidden max-h-full" class:hidden={sectionsHidden}>
		{#if containerEl}
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
			{#if collaborationState.provider}
				{#each sections as section, i}
					<div bind:this={sectionElements[section.field]}>
						{#if section.type == "timeline"}
							<IncidentTimeline {incident} />
						{:else if section.type === "field"}
							<EditorWrapper
								{section}
								provider={collaborationState.provider}
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
		<DiscussionSidebar bind:showDebrief retrospectiveId={retrospective.id} debriefId={debrief.id} />
	</div>
</div>

<Dialog
	bind:open={showDebrief}
	persistent
	portal
	classes={{ dialog: 'flex flex-col max-h-full w-5/6 max-w-7xl my-2', root: "p-2" }}
	>
	<div slot="header" class="border-b p-2" let:close>
		<Header title="Debrief">
			<svelte:fragment slot="actions">
				<Button on:click={() => close({force: true})} iconOnly icon={mdiClose} />
			</svelte:fragment>
		</Header>
	</div>

	<svelte:fragment slot="default">
		{#if showDebrief}
			<IncidentDebriefDialog {incident} {debrief} />
		{/if}
	</svelte:fragment>
</Dialog>