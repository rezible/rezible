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

	const getEditorOptionsDefinedElement = (e: Editor) => {
		let el = e.options.element;
		if (!el) return;
		if (typeof el === "function") return;
		if (el instanceof Element) return el; 
		return el.mount;
	}

	const setupEditor = () => {
		if (!editor?.options.element) return;
		if (editor.contentElement) return;

		const el = getEditorOptionsDefinedElement(editor);
		if (el) ref.append(...Array.from(el.childNodes));

		editor.setOptions({ element: ref });
		editor.contentElement = ref;
	};

	const destroyEditor = () => {
		if (!editor) return;

		editor.contentElement = null;

		const el = getEditorOptionsDefinedElement(editor);
		if (!el || !el.firstChild) return;

		const newRef = document.createElement("div");
		newRef.append(...Array.from(el.childNodes));
		editor.setOptions({ element: newRef });
	}

	onMount(() => {tick().then(setupEditor)});
	onDestroy(destroyEditor);
</script>

<div bind:this={ref} class={className}></div>

{#if children}{@render children()}{/if}
