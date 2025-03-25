<script lang="ts">
	import { Button, Header, MenuItem } from "svelte-ux";
	import { mdiFormatBold, mdiFormatListBulleted } from "@mdi/js";
	import { EditorContent } from "svelte-tiptap";
	import { handoverState, type HandoverEditorSection } from "../handover.svelte";
	import type { ChainedCommands } from "@tiptap/core";
	import {
		listOncallShiftAnnotationsOptions,
		listOncallShiftIncidentsOptions,
		type OncallShift,
		type OncallShiftHandover,
		type OncallShiftHandoverTemplate,
	} from "$lib/api";
	import { onMount } from "svelte";
	import { createQuery } from "@tanstack/svelte-query";

	type Props = {
		shift: OncallShift;
		template?: OncallShiftHandoverTemplate;
		handover?: OncallShiftHandover;
	};
	const { shift, template, handover }: Props = $props();

	const annotationsQuery = createQuery(() => listOncallShiftAnnotationsOptions({ path: { id: shift.id } }));
	const pinnedAnnotations = $derived(
		annotationsQuery.data?.data.filter((ann) => ann.attributes.pinned) ?? []
	);

	const incidentsEnabled = $derived(handoverState.sections.some((s) => s.kind === "incidents"));
	const incidentsQuery = createQuery(() => ({
		...listOncallShiftIncidentsOptions({ path: { id: shift.id } }),
		enabled: incidentsEnabled,
	}));
	const incidents = $derived(incidentsQuery.data?.data ?? []);

	const runEditorCmd = (toggleFn: (cmd: ChainedCommands) => void) => {
		return () => {
			if (!handoverState.activeEditor) return;
			const chain = handoverState.activeEditor.chain().focus();
			toggleFn(chain);
			chain.run();
		};
	};

	onMount(() => {
		if (template) handoverState.setupTemplate(template);
		if (handover) handoverState.restoreExisting(handover);
		return () => handoverState.destroy();
	});

	let focusIdx = $state(-1);
	const onSectionFocus = (e: FocusEvent, idx: number, focus: boolean) => {
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
</script>

{#each handoverState.sections as section, i}
	<div class="flex flex-col p-2">
		{#if section.header}
			<div class="flex w-full gap-4 items-center">
				<Header title={section.header} classes={{ root: "w-full", container: "flex-1" }} />
			</div>
		{/if}

		<div class="p-2 border border-surface-content/10 bg-surface-200/50">
			{#if section.kind === "annotations"}
				{@render annotationsSection()}
			{:else if section.kind === "incidents"}
				{@render incidentsSection()}
			{:else if section.kind === "regular"}
				{@render regularSection(i, section)}
			{/if}
		</div>
	</div>
{/each}

{#snippet annotationsSection()}
	{#if pinnedAnnotations.length === 0}
		<div>
			<span>No Events</span>
			<span class="text-surface-content/50">(Pinned Annotations will be included here)</span>
		</div>
	{:else}
		<ul class="list-disc pl-5">
			{#each pinnedAnnotations as ann (ann.id)}
				<li>{ann.attributes.title}</li>
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
	{@const activeStatus = section.activeStatus}
	{@const isActive = handoverState.activeEditor == section.editor && focusIdx === idx}
	<div
		class="h-fit"
		class:cursor-text={!handoverState.sent}
		onfocusin={(e) => {
			onSectionFocus(e, idx, true);
		}}
		onfocusout={(e) => {
			onSectionFocus(e, idx, false);
		}}
		tabindex="-1"
		spellcheck="false"
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
				variant={isActive && activeStatus.get("bold") ? "fill" : "fill-light"}
				on:click={runEditorCmd((c) => c.toggleBold())}
			/>

			<Button
				icon={mdiFormatListBulleted}
				rounded={false}
				size="sm"
				disabled={!isActive}
				variant={isActive && activeStatus.get("bulletList") ? "fill" : "fill-light"}
				on:click={runEditorCmd((c) => c.toggleBulletList())}
			/>
		</div>

		<EditorContent editor={section.editor} />
	</div>
{/snippet}
