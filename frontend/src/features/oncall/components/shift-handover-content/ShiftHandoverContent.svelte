<script lang="ts">
	import { onMount } from "svelte";
	import { Button, Header } from "svelte-ux";
	import { createQuery } from "@tanstack/svelte-query";
	import { EditorContent } from "svelte-tiptap";
	import type { ChainedCommands } from "@tiptap/core";
	import { mdiFormatBold, mdiFormatListBulleted } from "@mdi/js";
	import {
		listIncidentsOptions,
		listOncallShiftAnnotationsOptions,
		type OncallShiftHandover,
		type OncallShiftHandoverTemplate,
	} from "$lib/api";
	import { handoverState, type HandoverEditorSection } from "./state.svelte";
	import SendButton from "./SendButton.svelte";

	type Props = {
		shiftId: string;
		editable: boolean;
		handover?: OncallShiftHandover;
		template?: OncallShiftHandoverTemplate;
	};
	const { shiftId, editable, handover, template }: Props = $props();

	const isSent = $derived(new Date(handover?.attributes.sentAt ?? 0).valueOf() > 0);

	let focusIdx = $state(-1);
	const onSectionFocus = (e: FocusEvent, idx: number, focus: boolean) => {
		if (!editable || isSent) return;
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

	const annotationsQuery = createQuery(() => ({
		...listOncallShiftAnnotationsOptions({ path: { id: shiftId } }),
		enabled: !isSent,
	}));
	const annotations = $derived(annotationsQuery.data?.data ?? []);
	const pinnedAnnotations = $derived(annotations.filter((ann) => ann.attributes.pinned));

	const incidentsSectionPresent = $derived(handoverState.sections.some((s) => s.kind === "incidents"));
	const incidentsQuery = createQuery(() => ({
		...listIncidentsOptions({
			query: {
				/* TODO: filter by shiftId */
			},
		}),
		enabled: !isSent && incidentsSectionPresent,
	}));
	const incidents = $derived(incidentsQuery.data?.data ?? []);

	onMount(() => {
		handoverState.setup(handover, template);
		return () => handoverState.destroy();
	});
</script>

<div class="flex flex-col gap-2 shrink overflow-y-auto">
	{#each handoverState.sections as section, i}
		<div class="flex flex-col p-2">
			{#if section.header}
				<div class="flex w-full gap-4 items-center">
					<Header title={section.header} classes={{ root: "w-full", container: "flex-1" }} />
				</div>
			{/if}

			<div class="p-2 border border-surface-content/10 bg-surface-200/50">
				{#if section.kind === "regular"}
					{@render regularSection(i, section)}
				{:else if section.kind === "annotations"}
					{@render annotationsSection()}
				{:else if section.kind === "incidents"}
					{@render incidentsSection()}
				{/if}
			</div>
		</div>
	{/each}
</div>

{#if !handoverState.sent && editable}
	<div class="w-full flex justify-end px-2">
		<SendButton {shiftId} />
	</div>
{/if}

{#snippet annotationsSection()}
	{#if pinnedAnnotations.length === 0}
		<div>
			{#if handoverState.sent}
				<span class="text-surface-content/80">No Annotations Included</span>
			{:else}
				<span class="text-surface-content/50">Pinned Annotations will be included here</span>
			{/if}
		</div>
	{:else}
		<ul class="list-disc pl-5">
			{#each pinnedAnnotations as ann (ann.id)}
				<li>{ann.attributes.event?.title || "title"}</li>
				<ul class="pl-5">
					<li>
						<span class="italic">{ann.attributes.notes}</span>
					</li>
				</ul>
			{/each}
		</ul>
	{/if}
{/snippet}

{#snippet incidentsSection()}
	{#if incidents.length === 0}
		<span>No Incidents</span>
	{:else}
		<ul class="list-disc pl-5">
			{#each incidents as inc (inc.id)}
				<li>
					<a class="link" href="/incidents/{inc.id}" target="_blank">{inc.attributes.title}</a>
				</li>
			{/each}
		</ul>
	{/if}
{/snippet}

{#snippet regularSection(idx: number, section: HandoverEditorSection)}
	{@const isActive = handoverState.activeEditor == section.editor && focusIdx === idx}
	{#if !section.editor}
		<span class="text-surface-content/80">N/A</span>
	{:else}
		<div
			class="h-fit"
			tabindex="-1"
			spellcheck="false"
			class:cursor-text={!handoverState.sent}
			onfocusin={(e) => onSectionFocus(e, idx, true)}
			onfocusout={(e) => onSectionFocus(e, idx, false)}
		>
			<div
				class="flex flex-row w-full items-center gap-2 h-fit"
				class:hidden={handoverState.sent}
				data-menu={idx}
			>
				<Button
					icon={mdiFormatBold}
					rounded={false}
					size="sm"
					disabled={!isActive}
					variant={isActive && section.activeStatus?.get("bold") ? "fill" : "fill-light"}
					on:click={runEditorCmd((c) => c.toggleBold())}
				/>

				<Button
					icon={mdiFormatListBulleted}
					rounded={false}
					size="sm"
					disabled={!isActive}
					variant={isActive && section.activeStatus?.get("bulletList") ? "fill" : "fill-light"}
					on:click={runEditorCmd((c) => c.toggleBulletList())}
				/>
			</div>

			<EditorContent editor={section.editor} class="p-2" />
		</div>
	{/if}
{/snippet}
