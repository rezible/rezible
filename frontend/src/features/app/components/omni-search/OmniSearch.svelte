<script lang="ts">
	import { mdiMagnify } from "@mdi/js";

	import {
		type MenuOption,
		Button,
		Kbd,
		Dialog,
		SelectField,
	} from "svelte-ux";
	import { autoFocus, selectOnFocus } from '@layerstack/svelte-actions';
	import { smScreen } from '@layerstack/svelte-stores';
	import { cls } from '@layerstack/tailwind';
	import { searchState as search } from "./omni-search.svelte";
	import { goto } from "$app/navigation";

	let open = $state(false);

	const closeSearch = () => {
		open = false;
		search.clear();
	};

	const onSearchClicked = () => {
		open = true;
	};

	const onKeyDown = (e: KeyboardEvent) => {
		if (open && e.key === "Escape") {
			e.preventDefault();
			closeSearch();
			return;
		}
		const typeResult = search.isSearchKeyPress(e);
		if (!typeResult) return;
		e.preventDefault();
		open = true;
		search.startSearch(typeResult);
	};

	const fieldActions = (node: any) => [autoFocus(node), selectOnFocus(node)];

	const onInputChange = (value: string) => {
		search.updateInput(value);
	};

	const onSelected = (value?: MenuOption<string> | null) => {
		console.log(value);
		goto("/teams/sit");
		closeSearch();
	};

	const onDialogClose = () => {
		closeSearch();
	};

	const isGeneral = $derived(search.searchType === "general");
	const title = $derived(isGeneral ? "Search For Anything" : "Oncall Search");
	const placeholder = $derived(
		isGeneral ? "An Incident, Retrospective, Team Name..." : "Oncall Roster, Service, Username..."
	);
</script>

<svelte:window onkeydown={onKeyDown} />

<Button
	icon={mdiMagnify}
	iconOnly={!$smScreen}
	on:click={onSearchClicked}
	class={cls(
		"sm:bg-surface-100/60 sm:hover:bg-surface-100/80 text-surface-content/80 hover:text-surface-content rounded h-10 w-full max-w-xl mx-auto justify-start"
	)}
>
	<span class="flex-1 text-left max-sm:hidden">Search</span>
	<Kbd variant="none" class="opacity-75 max-sm:hidden" command>K</Kbd>
</Button>

<Dialog
	{open}
	on:close={onDialogClose}
	classes={{
		root: cls("items-start mt-8 sm:mt-24 sm:ml-16"),
		backdrop: "backdrop-blur-sm",
		title: "bg-surface-200 text-secondary-600",
	}}
>
	<div slot="title">
		<span>{title}</span>
	</div>

	<SelectField
		icon={mdiMagnify}
		inlineOptions={true}
		options={search.options}
		{placeholder}
		{fieldActions}
		on:inputChange={(e) => onInputChange(e.detail)}
		on:change={(e) => onSelected(e.detail.option)}
		classes={{
			root: "w-[420px] max-w-[95vw] py-1",
			field: {
				container: "border-none hover:shadow-none group-focus-within:shadow-none",
			},
			options: "overflow-auto max-h-[min(90dvh,380px)]",
			group: "capitalize",
		}}
	/>
</Dialog>
