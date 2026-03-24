<script lang="ts">
	import "./styles.postcss";
	import type { Editor as TiptapEditor } from "@tiptap/core";

	import { useIncidentView } from "$features/incidents/views/incident";

	import { useIncidentCollaboration } from "$features/incidents/views/incident/collaboration.svelte";
	import { draft } from "$features/incidents/views/incident/discussions.svelte";
	import { activeEditor } from "./activeEditor.svelte";

	import type { AnnotationType } from "./field-editor/BubbleMenu.svelte";
	import FieldEditorWrapper from "./field-editor/FieldEditorWrapper.svelte";
	import LoadingIndicator from "$src/components/loading-indicator/LoadingIndicator.svelte";

	const view = useIncidentView();
	const sections = $derived(view.retrospective?.attributes.reportSections ?? []);
	const collab = useIncidentCollaboration();

	let sectionsSidebarVisible = $state(false);

	let containerEl = $state<HTMLElement>();
	const sectionElements = $state<Record<string, HTMLElement>>({});

	const focusSectionFn = $state<Record<string, VoidFunction>>({});
	const onSectionClicked = (field: string) => {
		if (focusSectionFn[field]) focusSectionFn[field]();
		if (sectionElements[field]) sectionElements[field].scrollIntoView();
	};

	const onCreateAnnotation = (e: TiptapEditor, t: AnnotationType) => {
		if (t === "draft-comment") {
			draft.create(e);
		} else if (t === "service") {
		}
	};
</script>

<div class="flex flex-col min-h-0 overflow-y-auto grow">
	<div class="w-full overflow-y-auto flex flex-col gap-4" bind:this={containerEl}>
		{#if collab.provider}
			{#each sections as section, i}
				<div bind:this={sectionElements[section.field]}>
					<FieldEditorWrapper
						{section}
						provider={collab.provider}
						setIsActive={activeEditor.set}
						{onCreateAnnotation}
						bind:focusEditor={focusSectionFn[section.field]}
					/>
				</div>
			{/each}
		{:else}
			<span>Connecting...</span>
			<LoadingIndicator />
		{/if}
	</div>
</div>
