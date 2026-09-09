<script lang="ts">
	import { pageSizeOptions, QueryPaginator, type PaginatedQuery } from "$lib/api/queryPaginator.svelte";
	import * as Pagination from "$components/ui/pagination";
	import * as Select from "$components/ui/select";
	import type { Snippet } from "svelte";

	type Props = {
		paginator: QueryPaginator;
		query: PaginatedQuery;
		dense?: boolean;
		children: Snippet;
	};
	const { paginator, query, dense = false, children }: Props = $props();

	const fetching = $derived(query.isFetching);
	const pagination = $derived(query.data?.pagination);
	const total = $derived(pagination?.total ?? 0);

	const page = $derived(paginator.page);
	const pageSize = $derived(paginator.pageSize);
	const rangeStart = $derived(total === 0 ? 0 : (page - 1) * pageSize + 1);
	const rangeEnd = $derived(Math.min(page * pageSize, total));

	const onPageSizeSelected = (opt: string) => {
		if (opt) paginator.setPageSize(Number(opt));
	};

	const onPageChange = (newPage: number) => {
		paginator.setPage(newPage);
	};
</script>

<div class="flex min-h-0 flex-1 flex-col max-h-full" class:max-w-xl={dense}>
	<div class="flex flex-col gap-1 min-h-0 flex-1 overflow-auto pb-1 pr-1">
		{@render children()}
	</div>

	<div
		class="flex shrink-0 flex-wrap items-center justify-between gap-2 pt-2 text-sm text-muted-foreground"
	>
		<div class="flex items-center gap-2">
			<span>{rangeStart}–{rangeEnd} of {total}</span>
			{#if fetching}
				<span class="animate-pulse" aria-live="polite">Updating…</span>
			{/if}
		</div>

		<div class="flex items-center gap-3">
			<label class="flex items-center gap-2">
				<span class="hidden sm:inline">Rows per page</span>
				<Select.Root
					type="single"
					value={pageSize.toString()}
					onValueChange={onPageSizeSelected}
					disabled={fetching}
				>
					<Select.Trigger size="sm" aria-label="Rows per page">{pageSize}</Select.Trigger>
					<Select.Content>
						{#each pageSizeOptions as opt (opt)}
							<Select.Item value={opt.toString()} label={opt.toString()}>
								{opt}
							</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			</label>

			<Pagination.Root class="w-auto" count={total} perPage={pageSize} {page} {onPageChange}>
				{#snippet children({ pages, currentPage })}
					{@const isFirstPage = currentPage <= 1}
					{@const isLastPage = currentPage * pageSize >= total}
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.PrevButton disabled={fetching || isFirstPage} />
						</Pagination.Item>
						{#each pages as page (page.key)}
							<Pagination.Item>
								{#if page.type === "ellipsis"}
									<Pagination.Ellipsis />
								{:else}
									<Pagination.Link
										{page}
										disabled={fetching}
										isActive={page.value === currentPage}
									/>
								{/if}
							</Pagination.Item>
						{/each}
						<Pagination.Item>
							<Pagination.NextButton disabled={fetching || isLastPage} />
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	</div>
</div>
