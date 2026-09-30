<script lang="ts">
	import { onDestroy } from "svelte";
	import { watch } from "runed";
	import { cn } from "$lib/utils";
	import TiptapEditor from "$components/tiptap-editor/TiptapEditor.svelte";
	import { createReadOnlyDocumentEditor } from "$components/tiptap-editor/editors";

	type Props = {
		/** Markdown source. Empty or whitespace-only renders nothing. */
		markdown: string;
		/** "report": 15px/24px, max 72ch. "compact": 14px/22px, no max width. */
		size?: "report" | "compact";
		/** Added to Markdown heading levels for the rendered tag. Read once, when the editor is created. */
		headingOffset?: 0 | 1 | 2;
		/** Accessible name for the document region. */
		label: string;
		class?: string;
	};

	const { markdown, size = "report", headingOffset = 2, label, class: className }: Props = $props();

	// The editor is created once; later markdown changes are applied with setContent.
	// svelte-ignore state_referenced_locally
	const editor = createReadOnlyDocumentEditor(markdown, { headingOffset });
	// svelte-ignore state_referenced_locally
	let appliedMarkdown = markdown;

	const hasContent = $derived(markdown.trim().length > 0);

	watch(
		() => markdown,
		(next) => {
			if (next === appliedMarkdown) {
				return;
			}
			appliedMarkdown = next;
			editor.commands.setContent(next, { contentType: "markdown", emitUpdate: false });
		},
		{ lazy: true }
	);

	onDestroy(() => {
		editor.destroy();
	});
</script>

<div
	role="document"
	aria-label={label}
	data-size={size}
	class={cn("rich-text", !hasContent && "hidden", className)}
>
	<TiptapEditor {editor} />
</div>

<style>
	.rich-text {
		color: var(--foreground);
		overflow-wrap: anywhere;
	}

	.rich-text[data-size="report"] {
		font-size: 15px;
		line-height: 24px;
		max-width: 72ch;
	}

	.rich-text[data-size="compact"] {
		font-size: 14px;
		line-height: 22px;
	}

	.rich-text :global(.ProseMirror) {
		outline: none;
		white-space: pre-wrap;
		cursor: default;
	}

	.rich-text :global(.ProseMirror > * + *) {
		margin-top: 0.75em;
	}

	.rich-text :global([data-level="1"]) {
		font-size: 20px;
		line-height: 28px;
		font-weight: 600;
		margin-top: 1.5em;
	}

	.rich-text :global([data-level="2"]) {
		font-size: 18px;
		line-height: 26px;
		font-weight: 600;
		margin-top: 1.5em;
	}

	.rich-text :global([data-level="3"]) {
		font-size: 15px;
		line-height: 22px;
		font-weight: 600;
		margin-top: 1.25em;
	}

	.rich-text :global(:is([data-level="4"], [data-level="5"], [data-level="6"])) {
		font-size: 14px;
		line-height: 22px;
		font-weight: 600;
		margin-top: 1em;
	}

	.rich-text :global(.ProseMirror > :first-child) {
		margin-top: 0;
	}

	.rich-text :global(ul) {
		padding-inline-start: 1.25em;
		list-style: disc;
	}

	.rich-text :global(ol) {
		padding-inline-start: 1.25em;
		list-style: decimal;
	}

	.rich-text :global(li + li),
	.rich-text :global(li > :is(ul, ol)) {
		margin-top: 0.25em;
	}

	.rich-text :global(strong) {
		font-weight: 600;
	}

	.rich-text :global(code) {
		font-family: var(--font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
		font-size: 0.9em;
		background: var(--muted);
		border-radius: 4px;
		padding: 0.1em 0.3em;
	}

	.rich-text :global(pre) {
		font-family: var(--font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
		font-size: 13px;
		line-height: 20px;
		background: var(--muted);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 12px;
		overflow-x: auto;
	}

	.rich-text :global(pre code) {
		background: none;
		padding: 0;
		font-size: inherit;
	}

	.rich-text :global(blockquote) {
		border-inline-start: 2px solid var(--border);
		padding-inline-start: 12px;
		color: var(--muted-foreground);
	}

	.rich-text :global(a) {
		color: var(--foreground);
		text-decoration: underline;
		text-underline-offset: 3px;
		border-radius: 2px;
	}

	.rich-text :global(a:hover) {
		color: var(--foreground-emphasis);
	}

	.rich-text :global(a:focus-visible) {
		outline: 2px solid var(--ring);
		outline-offset: 2px;
	}

	.rich-text :global(hr) {
		border: 0;
		border-top: 1px solid var(--border);
		margin: 1.5em 0;
	}

	.rich-text :global(table) {
		display: block;
		overflow-x: auto;
		border-collapse: collapse;
		font-size: 14px;
		line-height: 20px;
	}

	.rich-text :global(:is(th, td)) {
		border-bottom: 1px solid var(--border);
		padding: 6px 12px 6px 0;
		text-align: start;
		vertical-align: top;
	}

	.rich-text :global(th) {
		font-weight: 500;
		color: var(--muted-foreground);
	}

	.rich-text :global(:is(th, td) p) {
		margin: 0;
	}
</style>
