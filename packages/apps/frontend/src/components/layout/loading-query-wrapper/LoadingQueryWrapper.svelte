<script lang="ts" generics="QueryData">
	import type { Snippet } from "svelte";
	import type { CreateQueryResult } from "@tanstack/svelte-query";
	import type { ErrorModel } from "$lib/api";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";

	type Props = {
		query: CreateQueryResult<{ data: QueryData }, ErrorModel>;
		view?: Snippet<[QueryData]>;
		loading?: Snippet;
		error?: Snippet<[ErrorModel]>;
		feedbackOnly?: boolean;
	};
	const { query, view, loading, error, feedbackOnly = false }: Props = $props();
</script>

{#if query.isPending && !query.data}
	{#if loading}
		{@render loading()}
	{:else}
		<LoadingIndicator />
	{/if}
{:else if query.isError && !query.data}
	<div role="alert" class="space-y-2">
		{#if error}
			{@render error(query.error as ErrorModel)}
		{:else}
			<InlineAlert error={query.error} />
		{/if}
		<Button variant="outline" size="sm" onclick={() => query.refetch()}>Try again</Button>
	</div>
{:else}
	{#if query.isFetching}
		<p role="status" class="px-3 text-xs text-muted-foreground">Refreshing…</p>
	{/if}
	{#if query.isError && query.data}
		<div role="alert" class="space-y-2 p-3">
			<p class="text-sm text-destructive">Refresh failed. Showing previously loaded data.</p>
			<InlineAlert error={query.error} />
			<Button variant="outline" size="sm" onclick={() => query.refetch()}>Try again</Button>
		</div>
	{/if}
	{#if !feedbackOnly && query.data && view}
		{@render view(query.data.data)}
	{/if}
{/if}
