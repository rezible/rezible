<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listSystemAnalysisComponentsOptions, type SystemAnalysisComponent } from "$lib/api";

	import { incidentCtx } from "$features/incidents/lib/context";
	import { Button, Field, Icon, ListItem, State, ToggleGroup, ToggleOption } from "svelte-ux";
	import { mdiPlus, mdiShapeSquareRoundedPlus } from "@mdi/js";
	import { getIconForComponentKind } from "$lib/systemComponents";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";

	const analysisId = incidentCtx.get().attributes.system_analysis_id;

	let selectingComponent = $state(false);
	let selectedComponent = $state<SystemAnalysisComponent>();
	const componentAttrs = $derived(selectedComponent?.attributes.component.attributes);

	const componentsQuery = createQuery(() => ({
		...listSystemAnalysisComponentsOptions({ path: { id: analysisId } }),
		enabled: selectingComponent,
	}));
	const incidentComponents = $derived(componentsQuery.data?.data ?? []);

	const selectComponent = (c: SystemAnalysisComponent) => (selectedComponent = c);
	const clearComponent = () => (selectedComponent = undefined);
	const onCancel = () => {
		if (!selectedComponent) selectingComponent = false;
		selectedComponent = undefined;
	}
	const confirmAddingComponent = () => {
		if (!selectedComponent) return;
		clearComponent();
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#if selectingComponent}
		<div class="border rounded flex flex-col gap-2 p-2">
			{#if !selectedComponent}
				{#each incidentComponents as c (c.id)}
					{@const attr = c.attributes.component.attributes}
					<ListItem
						title={attr.name}
						subheading={attr.description}
						avatar={{ class: "bg-surface-content/50 text-surface-100/90" }}
						class="flex-1"
						noShadow
					>
						<div slot="avatar" class="rounded-xl size-8 grid place-content-center">
							<Icon data={getIconForComponentKind(attr.kind)} classes={{root: "size-5"}} />
						</div>
						<div slot="actions">
							<Button icon={mdiShapeSquareRoundedPlus} iconOnly on:click={() => {selectComponent($state.snapshot(c))}} />
						</div>
					</ListItem>
				{/each}
				
				{#if incidentComponents.length === 0 && componentsQuery.isFetched}
					<span>No components linked to this incident</span>
				{/if}
			{:else}
				<span class="text-lg">{componentAttrs?.name}</span>

				<div class="flex flex-col gap-1">
					<span>Constraints</span>
					
					{#each (componentAttrs?.constraints ?? []) as cstr}
						<ListItem
							title={cstr.attributes.label}
							subheading={cstr.attributes.description}
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
				</div>
			{/if}
			<div class="w-full flex justify-end">
				<ConfirmButtons
					closeText="Cancel"
					onClose={onCancel}
					confirmText="Add"
					onConfirm={confirmAddingComponent}
					saveEnabled={!!selectedComponent}
				/>
			</div>
		</div>
	{:else}
		<!-- TODO: show context components -->

		<Button
			class="text-surface-content/50 p-2"
			color="primary"
			variant="fill-light"
			on:click={() => (selectingComponent = true)}
		>
			<span class="flex items-center gap-2 text-primary-content">
				Add Component
				<Icon data={mdiPlus} />
			</span>
		</Button>
	{/if}
</div>
