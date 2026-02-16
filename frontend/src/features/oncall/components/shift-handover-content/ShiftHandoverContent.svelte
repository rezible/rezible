<script lang="ts">
	import { mdiFormatBold, mdiFormatListBulleted } from "@mdi/js";
	import { ShiftHandoverEditorState, type HandoverEditorSection } from "./state.svelte";
	import TiptapEditor from "$components/tiptap-editor/TiptapEditor.svelte";
	import type { ChainedCommands } from "@tiptap/core";
	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";

	type Props = {
		handoverState: ShiftHandoverEditorState;
	};
	const { handoverState }: Props = $props();

	let focusIdx = $state(-1);
	const onSectionFocus = (e: FocusEvent, idx: number, focus: boolean) => {
		if (!handoverState.editable) return;
		handoverState.setEditorFocus(idx, focus);
		if (focus) focusIdx = idx;
		if (!focus && focusIdx === idx) {
			if (e.relatedTarget) {
				const btn = e.relatedTarget as HTMLButtonElement;
				const isMenuButton = btn.parentElement?.dataset["menu"] == String(idx);
				if (isMenuButton) return;
			}
			focusIdx = -1;
		}
	};

	const runEditorCmd = (toggleFn: (cmd: ChainedCommands) => void) => {
		return () => {
			if (!handoverState.activeEditor) return;
			const chain = handoverState.activeEditor.chain().focus();
			toggleFn(chain);
			chain.run();
		};
	};
</script>

<div class="flex flex-col gap-2 shrink overflow-y-auto">
	{#each handoverState.sections as section, i}
		{#if section.kind === "regular"}
			<div class="flex flex-col p-2">
				{#if section.header}
					<div class="flex w-full gap-4 items-center">
						<Header title={section.header} classes={{ root: "w-full", container: "flex-1" }} />
					</div>
				{/if}

				<div class="p-2 border border-surface-content/10 bg-surface-200/50">
					{@render editorSection(i, section)}
				</div>
			</div>
		{/if}
	{/each}
</div>

{#snippet editorSection(idx: number, section: HandoverEditorSection)}
	{@const isActive = handoverState.activeEditor == section.editor && focusIdx === idx}
	{#if !section.editor}
		<span class="text-surface-content/80">N/A</span>
	{:else}
		<div
			class="h-fit"
			tabindex="-1"
			spellcheck="false"
			class:cursor-text={!handoverState.isSent}
			onfocusin={(e) => onSectionFocus(e, idx, true)}
			onfocusout={(e) => onSectionFocus(e, idx, false)}
		>
			<div
				class="flex flex-row w-full items-center gap-2 h-fit"
				class:hidden={handoverState.isSent}
				data-menu={idx}
			>
				<Button
					size="sm"
					disabled={!isActive}
					onclick={runEditorCmd((c) => alert("TODO: migrate this"))}
				>bold</Button> <!-- variant={isActive && section.activeStatus?.get("bold") ? "fill" : "fill-light"} -->

				<Button
					size="sm"
					disabled={!isActive}
					onclick={runEditorCmd((c) => alert("TODO: migrate this"))}
				>bold</Button> <!-- variant={isActive && section.activeStatus?.get("bulletList") ? "fill" : "fill-light"} -->
			</div>

			<TiptapEditor bind:editor={section.editor} class="p-2" />
		</div>
	{/if}
{/snippet}
