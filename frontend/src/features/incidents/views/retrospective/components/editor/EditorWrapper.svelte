<script lang="ts">
    import './styles.postcss';
    import type { Editor as TiptapEditor } from '@tiptap/core';
    import { cls } from 'svelte-ux';
    import type { Retrospective } from "$lib/api";
    
    import { draft } from '$features/incidents/views/retrospective/lib/discussions.svelte';
    import { activeEditor } from '$features/incidents/views/retrospective/lib/editor.svelte';
    import { collaborationState } from '$features/incidents/views/retrospective/lib/collaboration.svelte';
    import IncidentTimeline from '$features/incidents/components/incident-timeline/IncidentTimeline.svelte';
    import type { AnnotationType } from './BubbleMenu.svelte';
    import SectionsSidebar from './SectionsSidebar.svelte';
    import FieldEditorWrapper from './FieldEditorWrapper.svelte';

    type Props = {
        incidentId: string;
        retrospective?: Retrospective;
    }
    const { incidentId, retrospective }: Props = $props();
    
	const sections = $derived(retrospective?.attributes.sections);

	let sectionsSidebarVisible = $state(false);

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

<div class="col-span-1 grow block overflow-y-hidden max-h-full" class:hidden={!sectionsSidebarVisible}>
    {#if sections && containerEl}
        <SectionsSidebar 
            bind:visible={sectionsSidebarVisible}
            {containerEl} {sections} {sectionElements} 
            {onSectionClicked} 
        />
    {/if}
</div>

<div class={cls(
    "flex flex-col grow min-h-0 overflow-y-auto bg-surface-300", 
    sectionsSidebarVisible ? "col-span-5" : "col-start-1 col-span-6 border-3"
)}>
    <div class="w-full overflow-y-auto pb-2 px-4 flex flex-col gap-4" bind:this={containerEl}>
        {#if sections && collaborationState.provider}
            {#each sections as section, i}
                <div bind:this={sectionElements[section.field]}>
                    {#if section.type == "timeline"}
                        <IncidentTimeline {incidentId} />
                    {:else if section.type === "field"}
                        <FieldEditorWrapper
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