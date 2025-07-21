<script lang="ts" module>
	import { Editor as TiptapEditor } from "@tiptap/core";

	export class Editor extends TiptapEditor {
		public contentElement: HTMLElement | null = null;
	}
</script>

<script lang="ts">
	import { onMount, onDestroy, tick, type Snippet } from "svelte";

	type Props = {
		editor: Editor;
		class?: string;
		children?: Snippet;
	};

	const { editor = $bindable(), children, class: className }: Props = $props();

	let ref = $state<HTMLElement>(null!);

	const setupEditor = () => {
		if (!editor?.options.element) return;
		if (editor.contentElement) return;

		ref.append(...Array.from(editor.options.element.childNodes));
		editor.setOptions({ element: ref });
		editor.contentElement = ref;
	};

	const destroyEditor = () => {
		if (!editor) return;

		editor.contentElement = null;

		if (!editor.options.element.firstChild) return;

		const newRef = document.createElement("div");
		newRef.append(...Array.from(editor.options.element.childNodes));
		editor.setOptions({ element: newRef });
	}

	onMount(() => {
		tick().then(setupEditor)
	});
	onDestroy(destroyEditor);
</script>

<div bind:this={ref} class={className}></div>

{#if children}{@render children()}{/if}
