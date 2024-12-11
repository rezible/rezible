<script lang="ts" module>
	interface NamedApiDataObject {
		id: string;
		attributes: {name: string};
	}
</script>

<script lang="ts" generics="DataType extends NamedApiDataObject">
	import { type ListQueryParameters, type ListQueryOptionsFunc } from '$lib/api';
	import {
		createQuery,
	} from '@tanstack/svelte-query';
	import { MultiSelectField, SelectField, type MenuOption } from 'svelte-ux';

	type Props = {
		options: ListQueryOptionsFunc<DataType>;
		value: string | string[] | undefined;
		id: string;
		label?: string | undefined;
		multi?: boolean;
		placeholder?: string | undefined;
		disabled?: boolean;
		clearable?: boolean;
	};
	let { 
		options,
		id,
		value = $bindable(),
		label,
		multi = false,
		placeholder,
		disabled = false,
		clearable = false,
	}: Props = $props();

	const params = $state<ListQueryParameters>({});
	const query = createQuery(() => options({query: params}));
	const data = $derived(query.data?.data ?? []);
	const canLoadMore = $derived((query.data?.pagination.total ?? 0) >= data.length);
	const loadedOptions = $derived<MenuOption<string>[]>(data.map((v) => ({ label: v.attributes.name, value: v.id })));

	const onSearchChange = async (input: string) => {
		// console.log(input);
	};
</script>

{#if query.error}
	<span>error fetching options: {query.error.message}</span>
{:else if multi && typeof value !== 'string'}
	<MultiSelectField
		{id}
		{label}
		bind:value
		menuProps={{ inlineSearch: true }}
		options={loadedOptions}
		disabled={disabled || !query.isSuccess}
		loading={query.isPending}
		{placeholder}
		on:change
	/>
{:else}
	<SelectField
		{id}
		{label}
		bind:value
		{clearable}
		options={loadedOptions}
		disabled={disabled || !query.isSuccess}
		loading={query.isPending}
		{placeholder}
		on:change
		search={onSearchChange}
	/>
{/if}
