<script lang="ts">
	import type { Snippet } from "svelte";
	import { appShell, type PageBreadcrumb } from "$features/app/lib/appShellState.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";

	type Props = {children: Snippet};
	const { children }: Props = $props();

	const pageBreadcrumbs = $derived(appShell.breadcrumbs);
	const pageActions = $derived(appShell.pageActions);
	const propsFn = $derived(pageActions?.propsFn ?? (() => ({})));
	const pageActionsProps = $derived.by(() => (propsFn()));
</script>

{#snippet breadcrumb(c: PageBreadcrumb, last: boolean)}
	<span class="flex items-stretch gap-2">
		{#if c.avatar}
			<Avatar {...c.avatar} size={30} />
		{/if}

		{#if c.href}
			<a href={c.href} class:text-3xl={last} class:text-surface-content={last}>{c.label ?? "loading"}</a>
		{:else}
			<span class:text-3xl={last} class:text-surface-content={last}>{c.label ?? "loading"}</span>
		{/if}
	</span>
{/snippet}

<div class="w-full max-w-full h-full max-h-full min-h-0 overflow-hidden flex flex-col gap-2 px-2 pb-2 border rounded-md bg-surface-200 text-surface-content">
	<div class="border-b flex justify-between items-bottom h-11">
		<span class="text-xl text-surface-content/50 w-fit self-bottom flex gap-1 items-end px-1">
			{#each pageBreadcrumbs as c, i}
				{#if i > 0}<span>/</span>{/if}
				{@render breadcrumb(c, i === pageBreadcrumbs.length - 1)}
			{/each}
		</span>

		{#if pageActions}
			<div class="flex items-center">
				<pageActions.component {...pageActionsProps} />
			</div>
		{/if}
	</div>

	<div class="flex flex-col flex-1 min-h-0 overflow-auto">
		{@render children()}
	</div>
</div>
