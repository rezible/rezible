<script lang="ts">
	import { Button, Header, MenuItem } from "svelte-ux";
	import { mdiFormatBold, mdiFormatListBulleted } from "@mdi/js";
	import { EditorContent } from "svelte-tiptap";
	import { handoverState } from "../handover.svelte";
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
	<div class="flex flex-col">
		{#if section.header}
			<div class="flex w-full gap-4 items-center">
				<Header title={section.header} classes={{ root: "w-full", container: "flex-1" }} />
			</div>
		{/if}

		{#if section.kind === "annotations"}
			<div class="bg-surface-100 p-2">
				{#if pinnedAnnotations.length === 0}
					<span
						>No Events <span class="text-surface-content/50"
							>(Pinned Annotations will be included here)</span
						></span
					>
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
			</div>
		{/if}

		{#if section.kind === "incidents"}
			<div class="bg-surface-100 p-2">
				{#if incidents.length === 0}
					<span>No Incidents</span>
				{:else}
					<ul class="list-disc pl-5">
						{#each incidents as inc (inc.id)}
							<li>
								<a class="link" href="/incidents/{inc.id}" target="_blank"
									>{inc.attributes.title}</a
								>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{/if}

		{#if section.kind === "regular"}
			{@const activeStatus = section.activeStatus}
			{@const isActive = handoverState.activeEditor == section.editor && focusIdx === i}
			<div
				class="bg-surface-100 p-1 h-fit"
				class:cursor-text={!handoverState.sent}
				onfocusin={(e) => {
					onSectionFocus(e, i, true);
				}}
				onfocusout={(e) => {
					onSectionFocus(e, i, false);
				}}
				tabindex="-1"
				spellcheck="false"
			>
				<div
					class="flex flex-row w-full items-center gap-2 h-fit"
					class:hidden={handoverState.sent}
					data-menu={i}
				>
					{#snippet formatMenuItem(name: string, active: boolean, icon: string, cmd: VoidFunction)}
						<MenuItem {icon} on:click={cmd} selected={active}>
							{name}
						</MenuItem>
					{/snippet}

					<!--Toggle let:on={open} let:toggle let:toggleOff>
						<Button
							disabled={!isActive}
							icon={isHeaderActive ? mdiFormatHeader1 : mdiText}
							on:click={toggle}
							classes={{ root: 'px-2' }}
							size="sm"
							variant={open ? 'fill' : 'fill-light'}
							rounded={false}
						>
							<Icon data={mdiChevronDown} />

							<Menu {open} on:close={toggleOff}>
								{@render formatMenuItem("Regular Text", !isHeaderActive, mdiText, runEditorCmd(c => c.setParagraph()))}
								{@render formatMenuItem("Heading", isHeaderActive, mdiFormatHeader1, runEditorCmd(c => c.toggleHeading({level: 1})))}
							</Menu>
						</Button>
					</Toggle-->

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
		{/if}
	</div>
{/each}
