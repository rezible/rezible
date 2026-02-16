<script lang="ts">
	import "./styles.postcss";
	import type { Editor as TiptapEditor } from "@tiptap/core";

	import { useIncidentViewController } from "$features/incidents/views/incident";

	import { draft } from "$src/features/incidents/lib/discussions.svelte";
	import { activeEditor } from "$src/features/incidents/lib/activeEditor.svelte";

	import type { AnnotationType } from "./field-editor/BubbleMenu.svelte";
	import FieldEditorWrapper from "./field-editor/FieldEditorWrapper.svelte";

	const view = useIncidentViewController();
	const sections = $derived(view.retrospective?.attributes.reportSections ?? []);
	const collab = $derived(view.collaboration);

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
			<span>connecting provider</span>
		{/if}
	</div>
</div>
