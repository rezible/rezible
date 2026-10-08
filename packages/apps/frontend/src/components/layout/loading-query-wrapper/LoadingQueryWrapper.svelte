<script lang="ts" generics="QueryData">
	import type { Snippet } from "svelte";
	import type { CreateQueryResult } from "@tanstack/svelte-query";
	import type { ApiError } from "$lib/api";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";

	type Props = {
		query: CreateQueryResult<{ data: QueryData }, ApiError>;
		view?: Snippet<[QueryData]>;
		loading?: Snippet;
		/** Error content. In "quiet" feedback this is the one-line message beside Retry. */
		error?: Snippet<[ApiError]>;
		/** Rendered instead of `view` when `isEmpty` returns true. */
		empty?: Snippet;
		isEmpty?: (data: QueryData) => boolean;
		/** "quiet" drops the refreshing indicator and uses one-line error states, for polled previews. */
		feedback?: "full" | "quiet";
		feedbackOnly?: boolean;
	};
	const {
		query,
		view,
		loading,
		error,
		empty,
		isEmpty,
		feedback = "full",
		feedbackOnly = false,
	}: Props = $props();

	const showEmpty = $derived(!!empty && !!query.data && !!isEmpty?.(query.data.data));
</script>

{#snippet retryButton(variant: "outline" | "ghost")}
	<Button {variant} size="sm" onclick={() => query.refetch()}>
		{feedback === "quiet" ? "Retry" : "Try again"}
	</Button>
{/snippet}

{#if query.isPending && !query.data}
	{#if loading}
		{@render loading()}
	{:else}
		<LoadingIndicator />
	{/if}
{:else if query.isError && !query.data}
	{#if feedback === "quiet"}
		<div role="alert" class="flex flex-wrap items-center gap-3 text-sm">
			{#if error}
				{@render error(query.error as ApiError)}
			{:else}
				<span>Could not load data.</span>
			{/if}
			{@render retryButton("outline")}
		</div>
	{:else}
		<div role="alert" class="space-y-2">
			{#if error}
				{@render error(query.error as ApiError)}
			{:else}
				<InlineAlert error={query.error} />
			{/if}
			{@render retryButton("outline")}
		</div>
	{/if}
{:else}
	{#if feedback === "quiet"}
		{#if query.isError && query.data}
			<div role="status" class="flex items-center gap-3 text-xs text-muted-foreground">
				<span>Refresh failed. Showing previously loaded data.</span>
				{@render retryButton("ghost")}
			</div>
		{/if}
	{:else}
		{#if query.isFetching}
			<p role="status" class="px-3 text-xs text-muted-foreground">Refreshing…</p>
		{/if}
		{#if query.isError && query.data}
			<div role="alert" class="space-y-2 p-3">
				<p class="text-sm text-destructive">Refresh failed. Showing previously loaded data.</p>
				<InlineAlert error={query.error} />
				{@render retryButton("outline")}
			</div>
		{/if}
	{/if}
	{#if !feedbackOnly && query.data}
		{#if showEmpty && empty}
			{@render empty()}
		{:else if view}
			{@render view(query.data.data)}
		{/if}
	{/if}
{/if}
