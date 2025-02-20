<script module lang="ts">
	interface ArchivableApiDataObject {
		id: string;
		attributes: {
			archived: boolean;
		};
	}
</script>

<script lang="ts" generics="DataType extends ArchivableApiDataObject">
	import type { Snippet } from "svelte";
	import {
		createMutation,
		createQuery,
		useQueryClient,
		type MutationOptions,
	} from "@tanstack/svelte-query";
	import { type ColumnDef } from "@layerstack/svelte-table";
	import { Button, Card, Pagination, Table } from "svelte-ux";
	import { paginationStore } from "@layerstack/svelte-stores";
	import type { PaginatedResponse, ErrorModel, ListQueryOptionsFunc, ListFuncQueryOptions } from "$lib/api";
	import { mdiArchive, mdiArchiveMinus, mdiPencil, mdiPlus } from "@mdi/js";
	import type { FormFields } from "./fields.svelte";
	import MutationForm from "./MutationForm.svelte";
	import ConfirmationModal from "./ConfirmationModal.svelte";

	type MutationOptionsFn = () => MutationOptions<any, ErrorModel, any>;
	type ObjectQueryOptions = {
		list: ListQueryOptionsFunc<DataType>;
		create: MutationOptionsFn;
		update: MutationOptionsFn;
		archive: MutationOptionsFn;
	};

	type Props = {
		fields: FormFields;
		dataType: string;
		description: string;
		headers: string[];
		listArchived?: boolean;
		dataRow: Snippet<[DataType]>;
		queryOptions: ObjectQueryOptions;
	};
	let {
		fields,
		dataType,
		description,
		headers,
		listArchived = true,
		dataRow,
		queryOptions,
	}: Props = $props();

	const queryClient = useQueryClient();
	const listParams = $state<ListFuncQueryOptions["query"]>({
		archived: listArchived,
	});
	const listQueryOptions = () => queryOptions.list({ query: listParams });

	let creating = $state(false);
	let archiveItem = $state<DataType>();
	let editItem = $state<DataType>();

	const archiveItemName = $derived(
		!!archiveItem && "name" in archiveItem ? `${dataType} '${archiveItem.name}` : dataType
	);

	const query = createQuery(listQueryOptions);
	const invalidateQuery = () => queryClient.invalidateQueries(listQueryOptions());

	const onMutationSuccess = () => {
		// TODO: optimistic update
		// queryClient.setQueryData(queryKey, (data: NamedArchivableApiDataObject[]) => {
		// 	return data;
		// });
		invalidateQuery();
	};

	const onMutationSettled = () => {
		archiveItem = undefined;
		archiveLoading = false;
	};

	const updateMutation = createMutation(() => ({
		...queryOptions.update(),
		onSuccess: onMutationSuccess,
		onSettled: onMutationSettled,
	}));
	const archiveMutation = createMutation(() => ({
		...queryOptions.archive(),
		onSuccess: onMutationSuccess,
		onSettled: onMutationSettled,
	}));

	let archiveLoading = $derived(updateMutation.isPending || archiveMutation.isPending);
	const toggleArchival = async () => {
		if (!archiveItem || archiveLoading) return;
		if (archiveItem.attributes.archived) {
			updateMutation.mutate({
				path: { id: archiveItem.id },
				body: { attributes: { archived: false } },
			});
		} else {
			archiveMutation.mutate({ path: { id: archiveItem.id } });
		}
	};

	const onEditSuccess = (item: DataType) => {
		editItem = undefined;
		creating = false;
		const opts = listQueryOptions();
		queryClient.setQueryData(
			opts.queryKey,
			(res: PaginatedResponse<DataType>): PaginatedResponse<DataType> => {
				if (!res) {
					return { data: [item], pagination: { total: 1 } };
				}
				const newRes = structuredClone(res);
				const index = res.data.findIndex((v) => v.id === item.id);
				if (index > -1) {
					// item exists in current data
					newRes.data[index] = item;
					return newRes;
				}

				console.log(item);

				// if (res.pagination.total < (listParams.limit ?? defaultListQueryLimit)) {
				// 	newRes.data = [...newRes.data, item];
				// 	newRes.pagination.total += 1;
				// } else {
				// 	newRes.data = [item];
				// 	newRes.pagination.total = 1;
				// 	// TODO: get current page?
				// 	newRes.pagination.previous = "prev";
				// }

				return newRes;
			}
		);
		queryClient.invalidateQueries(opts).then(() => {
			console.log("invalidated");
		});
	};

	const setEditItem = (item?: DataType) => {
		if (item && creating) {
			if (!confirm(`Cancel creating new ${dataType}?`)) return;
		}
		if (item && !!editItem) {
			if (!confirm(`Cancel editing current ${dataType}?`)) return;
		}
		editItem = item;
	};
	const editMutation = createMutation(() => ({
		...queryOptions.update(),
		onSuccess: ({ data }) => onEditSuccess(data),
		throwOnError: true,
	}));

	const setCreating = (isCreating: boolean) => {
		if (isCreating && !!editItem) {
			if (!confirm(`Cancel editing ${dataType}?`)) return;
			editItem = undefined;
		}
		creating = isCreating;
	};
	const createItemMutation = createMutation(() => ({
		...queryOptions.create(),
		onSuccess: ({ data }) => onEditSuccess(data),
		throwOnError: true,
	}));

	/*
	let searchValue: string | undefined = undefined;
	const debouncedSearch = () => {
		let timer: number;
		const debounceTime = 200;
		return () => {
			clearTimeout(timer);
            // @ts-ignore
			timer = setTimeout(() => {
				if (searchValue === '') searchValue = undefined;
				$listParams.search = searchValue;
			}, debounceTime);
		};
	};
    */

	const pagination = paginationStore();
	const updatePaginationOnQuery = (data?: PaginatedResponse<DataType>) => {
		pagination.setTotal(data?.pagination.total ?? 0);
	};
	// if (query.data) updatePaginationOnQuery(query.data);

	const columns: ColumnDef<DataType>[] = [
		...headers.map((name) => ({ name })),
		{ name: "actions", header: "", align: "right" },
	];
</script>

{#snippet mutatingTable(data: DataType[])}
	<Table {columns} {data} classes={{ container: "border p-2" }}>
		<tbody slot="data" let:data>
			{#each data ?? [] as row (row.id)}
				<tr>
					{#if editItem && editItem.id === row.id}
						<td colspan="2">
							<MutationForm
								data={row}
								{fields}
								{dataType}
								onMutate={(attributes) =>
									editMutation.mutate({
										path: { id: row.id },
										body: { attributes },
									})}
								isPending={editMutation.isPending}
								mutationError={editMutation.error}
								onClose={() => setEditItem(undefined)}
							/>
						</td>
					{:else}
						{@render mutatingTableRow(row)}
					{/if}
				</tr>
			{/each}
		</tbody>
	</Table>

	<Pagination
		{pagination}
		hideSinglePage
		show={["pagination", "prevPage", "nextPage"]}
		classes={{ perPage: "flex-1 text-right", pagination: "px-8" }}
	/>
{/snippet}

{#snippet mutatingTableRow(data: DataType)}
	{@render dataRow(data)}
	{@const disabled = creating || !!archiveItem || !!editItem}

	<td class="flex justify-end">
		{#if data.attributes.archived}
			<Button
				icon={mdiArchiveMinus}
				{disabled}
				on:click={() => {
					archiveItem = data;
				}}
			>
				Restore
			</Button>
		{:else}
			<Button icon={mdiPencil} {disabled} on:click={() => setEditItem(data)}>Edit</Button>
			<Button
				icon={mdiArchive}
				{disabled}
				on:click={() => {
					archiveItem = data;
				}}
			>
				Archive
			</Button>
		{/if}
	</td>
{/snippet}

<Card
	title="{dataType}s"
	subheading={description}
	class="p-4"
	classes={{ headerContainer: "px-0", content: "bg-surface-200" }}
>
	{#if query.isLoading}
		<span>Loading...</span>
	{:else if query.isError}
		<span>An error has occurred: {query.error}</span>
	{:else if query.isSuccess}
		{@const queryData = query.data.data}
		{#if queryData.length === 0}
			<span class:hidden={creating}>No {dataType}s!</span>
		{:else}
			{@render mutatingTable(queryData)}
		{/if}

		<div class="w-full mt-2" class:hidden={editItem || archiveItem}>
			{#if creating}
				<MutationForm
					{dataType}
					{fields}
					onMutate={(attributes) => createItemMutation.mutate({ body: { attributes } })}
					isPending={createItemMutation.isPending}
					mutationError={createItemMutation.error}
					onClose={() => setCreating(false)}
				/>
			{:else}
				<Button
					icon={mdiPlus}
					color="accent"
					variant="fill-outline"
					on:click={() => setCreating(true)}
				>
					Create New
				</Button>
			{/if}
		</div>
	{/if}
</Card>

<ConfirmationModal
	open={!!archiveItem}
	title="{archiveItem?.attributes.archived ? 'Restore' : 'Archive'} {dataType}"
	text="Are you sure you want to {archiveItem?.attributes.archived
		? 'restore'
		: 'archive'} the {archiveItemName}?"
	loading={archiveLoading}
	onConfirm={toggleArchival}
	onClose={() => {
		archiveItem = undefined;
	}}
/>
