<script lang="ts">
	import { mdiDotsVertical, mdiMenu, mdiPlus } from "@mdi/js";
	import {
		TextField,
		Button,
		Icon,
		type MenuOption,
		SelectField,
		MenuItem,
		cls,
		ListItem,
		MenuButton,
		Toggle,
		Menu,
	} from "svelte-ux";
	import ConfirmButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import { ContributingFactorCategories } from "../../types";
	import { add } from "date-fns";

	type EventFactor = {
		category: string;
		description: string;
	};
	let contributingFactors = $state<EventFactor[]>([]);

	type FactorMenuOption = MenuOption<string> & {
		description: string;
		examples: string;
	};
	const convertOptions = () => {
		let options: FactorMenuOption[] = [];
		ContributingFactorCategories.forEach((cat) => {
			const group = cat.name;
			cat.factors.forEach((factor) => {
				options.push({
					label: factor.title,
					value: factor.id,
					group,
					description: factor.description,
					examples: factor.examples.join(", "),
				});
			});
		});
		return options;
	};
	const options = convertOptions();

	let addingState = $state<EventFactor>();
	const selectedOption = $derived(
		addingState ? options.find((opt) => opt.value === addingState?.category) : undefined
	);

	const resetAddingState = () => {
		addingState = undefined;
	};
	const confirmAddingFactor = () => {
		if (!addingState) return;
		contributingFactors.push($state.snapshot(addingState));
		resetAddingState();
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#if addingState}
		<div class="flex flex-col gap-2 border rounded p-2">
			<SelectField bind:value={addingState.category} {options} label="Factor Type">
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

			{#if selectedOption}
				<TextField
					bind:value={addingState.description}
					label="Description"
					placeholder={selectedOption.description}
				></TextField>
			{/if}

			<div class="w-full flex justify-end">
				<ConfirmButtons
					onClose={resetAddingState}
					onConfirm={confirmAddingFactor}
					saveEnabled={!!selectedOption}
				/>
			</div>
		</div>
	{:else}
		{#each contributingFactors as fct, i}
			{@const categoryName =
				ContributingFactorCategories.find((c) => c.id === fct.category)?.name ?? "Unknown Category"}
			<ListItem
				title={fct.description}
				subheading={categoryName}
				classes={{ root: "pl-0" }}
				avatar={{ class: "bg-surface-content/50 text-surface-100/90" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions">
					<Toggle let:on={open} let:toggle let:toggleOff>
						<Button on:click={toggle}>
							<Icon data={mdiDotsVertical} />
							<Menu {open} on:close={toggleOff}>
								<MenuItem>Edit</MenuItem>
								<MenuItem>Remove</MenuItem>
							</Menu>
						</Button>
					</Toggle>
				</div>
			</ListItem>
		{/each}

		<Button
			class="text-surface-content/50 p-2"
			color="primary"
			variant="fill-light"
			on:click={() => {
				addingState = { category: "", description: "" };
			}}
		>
			<span class="flex items-center gap-2 text-primary-content">
				Add Factor
				<Icon data={mdiPlus} />
			</span>
		</Button>
	{/if}
</div>
