<script lang="ts" generics="QueryResultData">
	import type { Snippet } from "svelte";
	import type { CreateQueryResult } from "@tanstack/svelte-query";
	import { Card } from "svelte-ux";
	import { tryUnwrapApiError, type ErrorModel } from "$lib/api";
	import LoadingIndicator from "./LoadingIndicator.svelte";

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
		<Card
			title="Error {err.status} {err.title}"
			classes={{ header: { title: "text-danger text-xl" } }}
		>
			<div slot="contents" class="pb-3 flex flex-col">
				<span class="text-lg">{err.detail}</span>
				{#each err.errors ?? [] as d}
					<span
						>{d.location
							? `[${d.location}: "${d.value}"]: `
							: ""}<span class="text-neutral-content"
							>{d.message}</span
						></span
					>
				{/each}
			</div>
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
