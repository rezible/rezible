<script lang="ts" module>
	export type IncidentFieldOptions =
		| CreateIncidentFieldOptionAttributes[]
		| UpdateIncidentFieldOptionAttributes[];
</script>

<script lang="ts">
	import {
		type UpdateIncidentFieldOptionAttributes,
		type CreateIncidentFieldOptionAttributes,
		type fieldOptionType as OptionType,
	} from "$lib/api";
	import EditableListGroup from "$components/editable-list/EditableList.svelte";
	import type { EditorSnippetProps } from "$features/settings/components/mutating-table";
	import { SelectField, ToggleGroup, ToggleOption, type MenuOption } from "svelte-ux";

	type Props = EditorSnippetProps<IncidentFieldOptions>;
	const { id, value, onUpdate }: Props = $props();

	const creating = $derived(!value);
	const optionTypes: OptionType[] = ["custom", "derived"];
	let optionsType = $state<OptionType>("derived");

	type SelectCustomOptionType = MenuOption<string> & {
		label: string;
		archived: boolean;
	};
	let customOptions = $state<SelectCustomOptionType[]>([]);
	const sources: MenuOption<string>[] = [
		{ value: "services-names", label: "Services - Names" },
		{ value: "teams-names", label: "Teams - Names" },
	];
	let derivedFieldId = $state<string>();
	let derivationSource = $state<string>();

	const valueChanged = () => {
		let options: CreateIncidentFieldOptionAttributes[] | UpdateIncidentFieldOptionAttributes[] = [];
		if (optionsType === "custom") {
			options = customOptions.map((o) => ({
				id: o.value.length > 0 ? o.value : undefined,
				fieldOptionType: "custom",
				value: o.label,
				archived: creating ? undefined : o.archived,
			}));
		} else if (optionsType === "derived" && !!derivationSource) {
			if (creating) {
				options = [{ fieldOptionType: "derived", value: derivationSource }];
			} else {
				options = [
					{
						id: derivedFieldId,
						fieldOptionType: "derived",
						value: derivationSource,
						archived: false,
					},
				];
			}
		} else {
			console.log("invalid options type?");
			return;
		}
		// onUpdate(options);
	};

	const customOptionAdded = (e: CustomEvent) => {
		const val = e.detail as string;
		customOptions = [...customOptions, { label: val, value: "", archived: false }];
		console.log("add custom options");
		valueChanged();
	};

	const customOptionArchiveToggled = (e: CustomEvent) => {
		const idx = e.detail as number;
		if (!customOptions[idx]) return;
		if (creating) {
			customOptions.splice(idx - 1, 1);
		} else {
			customOptions[idx].archived = !customOptions[idx].archived;
		}
		customOptions = customOptions;
		valueChanged();
	};
</script>

<div class="flex flex-col gap-1">
	<span>label</span>

	<div class="border p-2">
		<div class="mb-2">
			<span>Type:</span>
			<ToggleGroup variant="outline" bind:value={optionsType} classes={{ root: "w-64" }}>
				{#each optionTypes as opt}
					<ToggleOption value={opt}>{opt}</ToggleOption>
				{/each}
			</ToggleGroup>
		</div>

		<div class:hidden={optionsType !== "custom"}>
			<EditableListGroup
				{id}
				bind:items={customOptions}
				deleteItems={creating}
				on:addItem={customOptionAdded}
				on:toggleArchived={customOptionArchiveToggled}
			/>
		</div>

		<div class:hidden={optionsType !== "derived"} class="">
			<SelectField
				{id}
				label="Source"
				options={sources}
				bind:value={derivationSource}
				on:change={valueChanged}
				placeholder="Select a data source for field options"
			/>
		</div>
	</div>
</div>
