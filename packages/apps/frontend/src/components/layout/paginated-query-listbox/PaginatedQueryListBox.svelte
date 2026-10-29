<script lang="ts">
	import RiArrowLeftSLine from "remixicon-svelte/icons/arrow-left-s-line";
	import RiArrowRightSLine from "remixicon-svelte/icons/arrow-right-s-line";
	import { pageSizeOptions, QueryPaginator, type PaginatedQuery } from "$lib/api/queryPaginator.svelte";
	import * as Pagination from "$components/ui/pagination";
	import * as Select from "$components/ui/select";
	import { cn } from "$lib/utils";
	import type { Snippet } from "svelte";

	type Props = {
		paginator: QueryPaginator;
		query: PaginatedQuery;
		/** Narrow box that scrolls its own rows, for lists embedded in a bounded panel. */
		dense?: boolean;
		/** Set false when the surrounding page already shows that results are refreshing. */
		showUpdating?: boolean;
		children: Snippet;
	};
	const { paginator, query, dense = false, showUpdating = true, children }: Props = $props();

	const fetching = $derived(query.isFetching);
	const pagination = $derived(query.data?.pagination);
	const total = $derived(pagination?.total ?? 0);

	const page = $derived(paginator.page);
	const pageSize = $derived(paginator.pageSize);
	const rangeStart = $derived(total === 0 ? 0 : (page - 1) * pageSize + 1);
	const rangeEnd = $derived(Math.min(page * pageSize, total));

	const showPages = $derived(total > pageSize || page > 1);
	const showPageSize = $derived(total > pageSizeOptions[0]);

	const onPageSizeSelected = (opt: string) => {
		if (opt) paginator.setPageSize(Number(opt));
	};

	const onPageChange = (newPage: number) => {
		paginator.setPage(newPage);
	};
</script>

<div class={cn("flex min-w-0 flex-col", dense && "max-h-full min-h-0 max-w-xl")}>
	<div class={cn("flex flex-col gap-1", dense && "min-h-0 flex-1 overflow-auto pr-1 pb-1")}>
		{@render children()}
	</div>

	{#if total > 0 || (fetching && showUpdating)}
		<div
			class="sticky bottom-0 z-10 flex shrink-0 flex-wrap items-center justify-between gap-2 bg-background py-3 text-sm text-muted-foreground"
		>
			<div class="flex items-center gap-2">
				{#if total > 0}
					<span class="tabular-nums">{rangeStart}–{rangeEnd} of {total}</span>
				{/if}
				{#if fetching && showUpdating}
					<span aria-live="polite">Updating…</span>
				{/if}
			</div>

			<div class="flex items-center gap-3">
				{#if showPageSize}
					<Select.Root
						type="single"
						value={pageSize.toString()}
						onValueChange={onPageSizeSelected}
						disabled={fetching}
					>
						<Select.Trigger size="sm" aria-label="Rows per page">
							<span class="tabular-nums">{pageSize} per page</span>
						</Select.Trigger>
						<Select.Content>
							{#each pageSizeOptions as opt (opt)}
								<Select.Item value={opt.toString()} label={`${opt} per page`}>
									{opt} per page
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				{/if}

				{#if showPages}
					<Pagination.Root class="w-auto" count={total} perPage={pageSize} {page} {onPageChange}>
						{#snippet children({ pages, currentPage })}
							{@const isFirstPage = currentPage <= 1}
							{@const isLastPage = currentPage * pageSize >= total}
							<Pagination.Content>
								<Pagination.Item>
									<Pagination.PrevButton
										class="size-8 p-0!"
										disabled={fetching || isFirstPage}
									>
										<RiArrowLeftSLine class="size-4" aria-hidden="true" />
									</Pagination.PrevButton>
								</Pagination.Item>
								{#each pages as page (page.key)}
									<Pagination.Item>
										{#if page.type === "ellipsis"}
											<Pagination.Ellipsis />
										{:else}
											<Pagination.Link
												class="size-8"
												{page}
												disabled={fetching}
												isActive={page.value === currentPage}
											/>
										{/if}
									</Pagination.Item>
								{/each}
								<Pagination.Item>
									<Pagination.NextButton
										class="size-8 p-0!"
										disabled={fetching || isLastPage}
									>
										<RiArrowRightSLine class="size-4" aria-hidden="true" />
									</Pagination.NextButton>
								</Pagination.Item>
							</Pagination.Content>
						{/snippet}
					</Pagination.Root>
				{/if}
			</div>
		</div>
	{/if}
</div>
