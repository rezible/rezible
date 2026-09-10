<script lang="ts">
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import { Button } from "$components/ui/button";
	import { v4 as uuidv4 } from "uuid";
	import ConfirmButtons from "$components/forms/confirm-buttons/ConfirmButtons.svelte";
	import { SvelteMap } from "svelte/reactivity";
	import { useEventDialogAttributes } from "./attributes.svelte";
	import type { TimelineEntryContributingFactor } from "../../entry-model";

	const attributes = useEventDialogAttributes();

	type FactorMenuOption = {
		label: string;
		value: string;
		group?: string;
		description: string;
		examples: string;
	};
	const factorTypeOptions: FactorMenuOption[] = [];

	const factorCategoryNames = $derived(
		new SvelteMap(factorTypeOptions.map((opt) => [opt.value, opt.group ?? "Unknown Category"]))
	);

	let editFactor = $state<TimelineEntryContributingFactor>();
	const selectedFactorType = $derived(
		editFactor
			? (factorTypeOptions.find((opt) => opt.value === editFactor?.attributes.factorTypeId) ?? {
					value: editFactor.attributes.factorTypeId,
					label: editFactor.attributes.factorTypeId,
					description: "",
					examples: "",
				})
			: undefined
	);

	const makeEmptyFactor = () => ({
		id: uuidv4(),
		attributes: { factorTypeId: "", description: "", links: [] },
	});

	const setEditing = (f?: TimelineEntryContributingFactor) => {
		editFactor = f ? $state.snapshot(f) : makeEmptyFactor();
	};

	const resetAddingState = () => {
		editFactor = undefined;
	};
	const confirmAddingFactor = () => {
		if (!editFactor) return;
		attributes.contributingFactors.push($state.snapshot(editFactor));
		resetAddingState();
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#if editFactor}
		<div class="flex flex-col gap-2 border rounded p-2">
			<span>type select</span>
			<!-- <SelectField
				bind:value={editFactor.attributes.factorTypeId}
				options={factorTypeOptions}
				loading={metadataQuery.isLoading}
				label="Factor Type"
			>
				<svelte:fragment slot="option" let:option let:index let:selected let:highlightIndex>
					<MenuItem
						on:click={() => {
							console.log("selected", option);
						}}
						class={cls(
							index === highlightIndex && "bg-surface-content/5",
							option === selected && "font-semibold",
							option.group ? "px-4" : "px-2"
						)}
						scrollIntoView={index === highlightIndex}
					>
						<div>
							<div>{option.label}</div>
							<div class="text-sm text-surface-content/50">
								{option.examples}
							</div>
						</div>
					</MenuItem>
				</svelte:fragment>
			</SelectField> -->

			{#if selectedFactorType}
				<span>description</span>
				<!-- <TextField
					bind:value={editFactor.attributes.description}
					label="Description"
					placeholder={editFactor.attributes.description}
				></TextField> -->
			{/if}

			<div class="w-full flex justify-end">
				<ConfirmButtons
					closeText="Cancel"
					onClose={resetAddingState}
					onConfirm={confirmAddingFactor}
					saveEnabled={!!selectedFactorType}
				/>
			</div>
		</div>
	{:else}
		{#each attributes.contributingFactors as f (f.id)}
			{@const categoryName = factorCategoryNames.get(f.attributes.factorTypeId) ?? "Unknown Category"}
			<span>factor: {categoryName}</span>
			<!-- <ListItem
				title={f.attributes.description}
				subheading={categoryName}
				classes={{ root: "pl-0" }}
				avatar={{ class: "bg-surface-content/50 text-surface-100/90" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions" class="flex gap-2 items-center">
					<Button icon={RiDeleteBinLine} iconOnly onclick={() => confirmRemoveFactor(f)} />
					<Button icon={RiPencilLine} iconOnly onclick={() => setEditing(f)} />
				</div>
			</ListItem> -->
		{/each}

		<Button color="primary" onclick={() => setEditing()}>
			<span class="flex items-center gap-2 text-primary-content">
				Add Factor
				<RiAddLine class="" aria-hidden="true" />
			</span>
		</Button>
	{/if}
</div>
