<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import {
		listSystemAnalysisComponentsOptions,
		type IncidentEventSystemComponent,
		type IncidentEventSystemComponentAttributes,
		type SystemAnalysisComponent,
	} from "$lib/api";
	import { v4 as uuidv4 } from "uuid";
	import { getIconForComponentKind } from "$lib/systemComponents";
	import { SvelteMap } from "svelte/reactivity";
	import { ListItem, TextField, ToggleGroup, ToggleOption } from "svelte-ux";
	import Button from "$components/button/Button.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiPencil, mdiPlus, mdiShapeSquareRoundedPlus, mdiTrashCan } from "@mdi/js";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";

	import { useIncidentAnalysis } from "../../../analysisState.svelte";
	import { eventAttributes } from "./eventAttributesState.svelte";

	const analysis = useIncidentAnalysis();
	const analysisId = $derived(analysis.analysisId);

	const analysisComponentsQuery = createQuery(() => ({
		...listSystemAnalysisComponentsOptions({ path: { id: analysisId } }),
		enabled: !!analysisId,
	}));
	const analysisComponents = $derived(analysisComponentsQuery.data?.data ?? []);
	const analysisComponentMap = $derived(
		new SvelteMap(analysisComponents.map((c) => [c.id, c]))
	);

	let description = $state("");
	let status = $state<string | "normal" | "degraded" | "failing">("normal");

	const getAttributes = (cmp: SystemAnalysisComponent): IncidentEventSystemComponentAttributes => ({
		analysisComponentId: $state.snapshot(cmp.id),
		description: $state.snapshot(description),
		status: $state.snapshot(status),
	});

	let selecting = $state(false);
	let selectedComponent = $state<SystemAnalysisComponent>();
	let editing = $state<IncidentEventSystemComponent>();
	const editComponent = $derived(
		editing ? analysisComponentMap.get(editing.attributes.analysisComponentId) : undefined
	);

	/*
	const focusComponentId = $derived(selectedComponent?.id ?? editComponent?.id);
	const relationshipsQuery = createQuery(() => ({
		...listSystemAnalysisRelationshipsOptions({ 
			path: { id: analysisId }, 
			query: { analysisComponentId: focusComponentId } }),
		enabled: !!focusComponentId,
	}));
	const relationships = $derived(!relationshipsQuery.isStale ? (relationshipsQuery.data?.data ?? []) : []);
	*/

	const setEditing = (cx: IncidentEventSystemComponent) => {
		editing = $state.snapshot(cx);
		description = $state.snapshot(cx.attributes.description);
		status = $state.snapshot(cx.attributes.status);
	};

	const confirmDelete = (cx: IncidentEventSystemComponent) => {
		const cmp = analysisComponentMap.get(cx.attributes.analysisComponentId);
		editing = undefined;
		if (!cmp || !confirm(`Are you sure you want to remove ${cmp.attributes.component.attributes.name}?`)) return;
		const idx = eventAttributes.systemContext.findIndex((c) => c.id === cx.id);
		if (idx >= 0) eventAttributes.systemContext.splice(idx, 1);
	};

	const resetState = () => {
		selecting = false;
		selectedComponent = undefined;
		editing = undefined;
		description = "";
		status = "normal";
	}

	const onCancel = () => {
		if (selecting && selectedComponent) {
			selectedComponent = undefined;
		} else {
			resetState();
		}
	};

	const onConfirm = () => {
		if (selecting && selectedComponent) {
			eventAttributes.systemContext.push({
				id: uuidv4(),
				attributes: getAttributes(selectedComponent),
			});
		} else if (editing && editComponent) {
			const idx = eventAttributes.systemContext.findIndex((c) => c.id === editing?.id);
			if (idx < 0) return;
			eventAttributes.systemContext[idx].attributes = getAttributes(editComponent);
		}
		resetState();
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#snippet componentContextEditor(cmp: SystemAnalysisComponent)}
		{@const attrs = cmp.attributes.component.attributes}
		<span class="text-lg">{attrs.name}</span>

		<ToggleGroup variant="fill-surface" bind:value={status} gap>
			<ToggleOption value="normal">Normal</ToggleOption>
			<ToggleOption value="degraded">Degraded</ToggleOption>
			<ToggleOption value="failing">Failing</ToggleOption>
		</ToggleGroup>

		<TextField label="Description" bind:value={description} multiline />

		<!-- TODO: think about what we want as rich context -->
		<!--div class="flex flex-col gap-1">
			<span>Relationships</span>

			{#each relationships ?? [] as rel}
				{@const attrs = rel.attributes}
				{@const otherId = attrs.sourceId === focusComponentId ? attrs.targetId : attrs.sourceId}
				{@const other = analysisComponentMap.get(otherId)?.attributes.component}
				<ListItem
					title={other?.attributes.name ?? "Unknown Component"}
					subheading={rel.attributes.description}
					classes={{ root: "border first:border-t rounded elevation-0" }}
					class="flex-1"
					noShadow
				>
					<div slot="actions">
						<ToggleGroup variant="fill-surface" value={"normal"} gap>
							<ToggleOption value="normal">Normal</ToggleOption>
							<ToggleOption value="degraded">Degraded</ToggleOption>
							<ToggleOption value="failing">Failing</ToggleOption>
						</ToggleGroup>
					</div>
				</ListItem>
			{/each}
		</div-->
	{/snippet}

	{#snippet confirmButtons()}
		<div class="w-full flex justify-end">
			<ConfirmButtons
				closeText="Cancel"
				onClose={onCancel}
				confirmText={selecting ? "Add" : "Save"}
				{onConfirm}
				saveEnabled={!!selectedComponent}
			/>
		</div>
	{/snippet}

	{#snippet componentSelector()}
		{#each analysisComponents as c (c.id)}
		{@const attr = c.attributes.component.attributes}
			<ListItem
				title={attr.name}
				subheading={attr.description}
				avatar={{ class: "bg-surface-content/50 text-surface-100/90" }}
				class="flex-1"
				noShadow
			>
				<div slot="avatar" class="rounded-xl size-8 grid place-content-center">
					<Icon data={getIconForComponentKind(attr.kindId)} classes={{ root: "size-5" }} />
				</div>
				<div slot="actions">
					<Button
						icon={mdiShapeSquareRoundedPlus}
						iconOnly
						on:click={() => (selectedComponent = $state.snapshot(c))}
					/>
				</div>
			</ListItem>
		{/each}

		{#if analysisComponents.length === 0 && analysisComponentsQuery.isFetched}
			<span>No components linked to this incident</span>
		{/if}
	{/snippet}

	{#if selecting || editing}
		<div class="border rounded flex flex-col gap-2 p-2">
			{#if selecting}
				{#if selectedComponent}
					{@render componentContextEditor(selectedComponent)}
				{:else}
					{@render componentSelector()}
				{/if}

				{@render confirmButtons()}
			{:else if editing}
				{#if editComponent}
					{@render componentContextEditor(editComponent)}
				{/if}

				{@render confirmButtons()}
			{/if}
		</div>
	{:else}
		{#each eventAttributes.systemContext as cx, i}
			{@const cmp = analysisComponentMap.get(cx.attributes.analysisComponentId)?.attributes.component}
			<ListItem
				title={cmp?.attributes.name ?? "Unknown Component"}
				subheading={cx.attributes.description}
				classes={{ root: "border first:border-t rounded elevation-0" }}
				class="flex-1"
				noShadow
			>
				<div slot="avatar">
					<span>{cx.attributes.status}</span>
				</div>
				<div slot="actions">
					<Button icon={mdiPencil} iconOnly on:click={() => setEditing(cx)} />
					<Button icon={mdiTrashCan} iconOnly on:click={() => confirmDelete(cx)} />
				</div>
			</ListItem>
		{/each}

		<Button
			class="text-surface-content/50 p-2"
			color="primary"
			variant="fill-light"
			on:click={() => (selecting = true)}
		>
			<span class="flex items-center gap-2 text-primary-content">
				Add Component
				<Icon data={mdiPlus} />
			</span>
		</Button>
	{/if}
</div>
