<script lang="ts">
    import './styles.postcss';
    import type { Editor as TiptapEditor } from '@tiptap/core';
    
    import { draft } from '$features/incidents/lib/discussions.svelte';
    import { activeEditor } from '$features/incidents/lib/editor.svelte';
	
    import type { AnnotationType } from './field-editor/BubbleMenu.svelte';
    import FieldEditorWrapper from './field-editor/FieldEditorWrapper.svelte';

	import type { Incident, Retrospective } from '$lib/api';
	import { collaborationState } from '$features/incidents/lib/collaboration.svelte';

	type Props = { 
		incident: Incident;
		retrospective: Retrospective;
	};
	let { incident, retrospective }: Props = $props();

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

<div class="flex flex-col min-h-0 overflow-y-auto bg-surface-300 grow">
    <div class="w-full overflow-y-auto pb-2 px-4 flex flex-col gap-4" bind:this={containerEl}>
        {#if sections && collaborationState.provider}
            {#each sections as section, i}
                <div bind:this={sectionElements[section.field]}>
                    <FieldEditorWrapper
                        {section}
						provider={collaborationState.provider}
                        setIsActive={activeEditor.set}
                        {onCreateAnnotation}
                        bind:focusEditor={focusSectionFn[section.field]} 
                    />
                </div>
            {/each}
        {:else}
            <span>connecting provider</span>
        {/if}
    </div>
</div>