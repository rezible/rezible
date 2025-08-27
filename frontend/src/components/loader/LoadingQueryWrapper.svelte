<script lang="ts" generics="QueryResultData">
	import type { Snippet } from "svelte";
	import type { CreateQueryResult } from "@tanstack/svelte-query";
	import type { ErrorModel } from "$lib/api";
	import Card from "$components/card/Card.svelte";
	import Header from "$components/header/Header.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";

	type Props = {
		query: CreateQueryResult<{ data: QueryResultData }, Error>;
		view: Snippet<[QueryResultData]>;
		loading?: Snippet;
		error?: Snippet<[ErrorModel]>;
	};
	const { query, view, loading, error }: Props = $props();
</script>

{#snippet defaultErrorView(err: ErrorModel)}
	{#if err.status}
		<Card classes={{ root: "mb-2" }}>
			{#snippet header()}
				<Header title="Error {err.status} {err.title}" classes={{ title: "text-danger text-xl" }} />
			{/snippet}
			{#snippet contents()}
				<div class="pb-3 flex flex-col">
					<span class="text-lg">{err.detail}</span>
					{#each err.errors ?? [] as d}
						<div>
							<span>{d.location ? `[${d.location}: "${d.value}"]: ` : ""}</span>
							<span class="text-neutral-content">{d.message}</span>
						</div>
					{/each}
				</div>
			{/snippet}
		</Card>
	{:else}
		<span>error: {query.error}</span>
	{/if}
{/snippet}

{#if query.isLoading}
	{#if loading}
		{@render loading()}
	{:else}
		<LoadingIndicator />
	{/if}
{:else if query.isError}
	{#if error}
		{@render error(query.error as ErrorModel)}
	{:else}
		{@render defaultErrorView(query.error as ErrorModel)}
	{/if}
{:else if query.isSuccess}
	{@render view(query.data.data)}
{/if}
