<script lang="ts" generics="CombinedQueryResultData">
	import { createQueries, type QueriesResults } from "@tanstack/svelte-query";
	import LoadingIndicator from "./LoadingIndicator.svelte";
	import type { Snippet } from "svelte";
	import { tryUnwrapApiError, type ErrorModel } from "$lib/api";
	import Card from "$components/card/Card.svelte";
	import Header from "$components/header/Header.svelte";

	// dont use this

	type Props = {
		queries: ReturnType<typeof createQueries<any>>;
		view: Snippet<[any]>;
		loading?: Snippet;
		error?: Snippet<[(ErrorModel | null)[]]>;
	};
	const { queries, view, loading, error }: Props = $props();

	const isLoading = $derived(!!queries.find((q) => q.isLoading));
	const isError = $derived(!!queries.find((q) => q.isError));
	const errors = $derived(queries.map((q) => (q.isError ? tryUnwrapApiError(q.error) : null)));
	const data = $derived(queries.map((q) => q.data ?? null));
</script>

{#snippet defaultErrorView(err: ErrorModel)}
	{#if err.status}
		<Card>
			{#snippet header()}
				<Header title="Error {err.status} {err.title}" classes={{ title: "text-danger text-xl" }} />
			{/snippet}
			{#snippet contents()}
				<div class="pb-3 flex flex-col">
					<span class="text-lg">{err.detail}</span>
					{#each err.errors ?? [] as d}
						<span
							>{d.location ? `[${d.location}: "${d.value}"]: ` : ""}<span
								class="text-neutral-content">{d.message}</span
							></span
						>
					{/each}
				</div>
			{/snippet}
		</Card>
	{:else}
		<span>error: {errors}</span>
	{/if}
{/snippet}

{#if isLoading}
	{#if loading}
		{@render loading()}
	{:else}
		<LoadingIndicator />
	{/if}
{:else if isError}
	{#if error}
		{@render error(errors)}
	{:else}
		{#each errors as err, i}
			{#if !!err}
				{@render defaultErrorView(err)}
			{/if}
		{/each}
	{/if}
{/if}

{@render view(data)}
