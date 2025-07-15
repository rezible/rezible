<script lang="ts">
	import type { Snippet } from "svelte";
	import { appShell, type PageBreadcrumb } from "$features/app/lib/appShellState.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { Button } from "svelte-ux";
	import { mdiDockLeft } from "@mdi/js";

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

<div class="w-full max-w-full h-full max-h-full min-h-0 overflow-hidden flex flex-col gap-2 text-surface-content">
	<!-- <div class="flex justify-between items-center h-11 border border-surface-content/10 rounded-md pl-2 pr-1"> -->
	<div class="flex justify-between items-center h-11 rounded-md">
		<div class="flex items-center gap-2">
			<Button icon={mdiDockLeft} iconOnly size="sm" classes={{root: "text-surface-content/40"}} />

			<span class="text-xl text-surface-content/50 w-fit self-bottom flex gap-1 items-end">
				{#each pageBreadcrumbs as c, i}
					{#if i > 0}<span>/</span>{/if}
					{@render breadcrumb(c, i === pageBreadcrumbs.length - 1)}
				{/each}
			</span>
		</div>

		{#if pageActions}
			<div class="flex items-center">
				<pageActions.component {...pageActionsProps} />
			</div>
		{/if}
	</div>

	<!-- <div class="flex flex-col flex-1 min-h-0 overflow-auto p-2 border border-surface-content/10 rounded-md bg-surface-200"> -->
	<div class="flex flex-col flex-1 min-h-0 overflow-auto">
		{@render children()}
	</div>
</div>