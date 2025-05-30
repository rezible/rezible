<script lang="ts">
	import { mdiPencil, mdiPlus, mdiTrashCan } from "@mdi/js";
	import {
		TextField,
		Button,
		type MenuOption,
		SelectField,
		MenuItem,
		ListItem,
	} from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { cls } from '@layerstack/tailwind';
	import { v4 as uuidv4 } from "uuid";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import {
		listIncidentEventContributingFactorCategoriesOptions,
		type IncidentEventContributingFactor,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { SvelteMap } from "svelte/reactivity";
	import { eventAttributes } from "./eventAttributesState.svelte";

	const categoriesQuery = createQuery(() => listIncidentEventContributingFactorCategoriesOptions());
	const categories = $derived(categoriesQuery.data?.data ?? []);

	type FactorMenuOption = MenuOption<string> & {
		description: string;
		examples: string;
	};
	const factorTypeOptions = $derived.by(() => {
		let options: FactorMenuOption[] = [];
		categories.forEach((cat) => {
			cat.attributes.factorTypes.forEach(({ id, attributes }) => {
				options.push({
					value: id,
					label: attributes.name,
					group: cat.attributes.name,
					description: attributes.description,
					examples: attributes.examples.join(", "),
				});
			});
		});
		return options;
	});

	const factorCategoryNames = $derived(
		new SvelteMap(factorTypeOptions.map((opt) => [opt.value, opt.group ?? "Unknown Category"]))
	);

	let editFactor = $state<IncidentEventContributingFactor>();
	const selectedFactorType = $derived(
		editFactor
			? factorTypeOptions.find((opt) => opt.value === editFactor?.attributes.factorTypeId)
			: undefined
	);

	const makeEmptyFactor = () => ({
		id: uuidv4(),
		attributes: { factorTypeId: "", description: "", links: [] },
	});

	const setEditing = (f?: IncidentEventContributingFactor) => {
		editFactor = f ? $state.snapshot(f) : makeEmptyFactor();
	};

	const confirmRemoveFactor = (f: IncidentEventContributingFactor) => {
		if (!confirm("Are you sure you want to remove this factor?")) return;
		const newFactors = eventAttributes.contributingFactors.filter(v => v.id !== f.id);
		eventAttributes.contributingFactors = newFactors;
	};

	const resetAddingState = () => {
		editFactor = undefined;
	};
	const confirmAddingFactor = () => {
		if (!editFactor) return;
		eventAttributes.contributingFactors.push($state.snapshot(editFactor));
		resetAddingState();
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#if editFactor}
		<div class="flex flex-col gap-2 border rounded p-2">
			<SelectField
				bind:value={editFactor.attributes.factorTypeId}
				options={factorTypeOptions}
				loading={categoriesQuery.isLoading}
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
			</SelectField>

			{#if selectedFactorType}
				<TextField
					bind:value={editFactor.attributes.description}
					label="Description"
					placeholder={editFactor.attributes.description}
				></TextField>
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
		{#each eventAttributes.contributingFactors as f}
			{@const categoryName = factorCategoryNames.get(f.attributes.factorTypeId) ?? "Unknown Category"}
			<ListItem
				title={f.attributes.description}
				subheading={categoryName}
				classes={{ root: "pl-0" }}
				avatar={{ class: "bg-surface-content/50 text-surface-100/90" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions" class="flex gap-2 items-center">
					<Button icon={mdiTrashCan} iconOnly on:click={() => confirmRemoveFactor(f)} />
					<Button icon={mdiPencil} iconOnly on:click={() => setEditing(f)} />
				</div>
			</ListItem>
		{/each}

		<Button
			class="text-surface-content/50 p-2"
			color="primary"
			variant="fill-light"
			on:click={() => setEditing()}
		>
			<span class="flex items-center gap-2 text-primary-content">
				Add Factor
				<Icon data={mdiPlus} />
			</span>
		</Button>
	{/if}
</div>
