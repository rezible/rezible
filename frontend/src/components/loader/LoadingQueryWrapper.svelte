<script lang="ts" generics="QueryResultData">
	import type { Snippet } from "svelte";
	import type { CreateQueryResult } from "@tanstack/svelte-query";
	import { tryUnwrapApiError, type ErrorModel } from "$lib/api";
	import LoadingIndicator from "./LoadingIndicator.svelte";
	import Card from "$components/card/Card.svelte";
	import Header from "$components/header/Header.svelte";

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
	{@const err = tryUnwrapApiError(query.error)}
	{#if error}
		{@render error(err)}
	{:else}
		{@render defaultErrorView(err)}
	{/if}
{/if}

{#if query.data}
	{@render view(query.data.data)}
{/if}
