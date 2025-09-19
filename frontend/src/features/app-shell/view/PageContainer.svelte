<script lang="ts">
	import type { Snippet } from "svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { useAuthSessionState } from "$lib/auth.svelte";

	type Props = {
		children: Snippet;
	};
	const { children }: Props = $props();

	const session = useAuthSessionState();

	const pageBreadcrumbs = $derived(appShell.breadcrumbs);
	const pageActions = $derived(appShell.pageActions);
	const propsFn = $derived(pageActions?.propsFn ?? (() => ({})));
	const pageActionsProps = $derived.by(() => (propsFn()));
</script>

{#snippet breadcrumbs()}
	<span class="w-fit self-bottom flex gap-2 items-end">
		{#each pageBreadcrumbs as c, i}
			{@const last = i === pageBreadcrumbs.length - 1}
			{#if i > 0}<span class="text-surface-content/50">/</span>{/if}
			<span class="flex items-stretch gap-2 text-surface-content/50 text-lg">
				{#if c.avatar}
					<span class="self-center"><Avatar {...c.avatar} size={24} /></span>
				{/if}

				<svelte:element this={c.href ? "a" : "span"} href={c.href} class={{"text-2xl": last, "text-surface-content/80": last}}>
					{c.label ?? ""}
				</svelte:element>
			</span>
		{/each}
	</span>
{/snippet}

<div class="flex justify-between items-center h-11 rounded-md bg-surface-200/80" class:hidden={!session.isSetup}>
	<div class="flex items-center gap-2 px-2">
		{@render breadcrumbs()}
	</div>

	{#if pageActions}
		<div class="flex items-center">
			<pageActions.component {...pageActionsProps} />
		</div>
	{/if}
</div>

<div class="flex flex-col flex-1 min-h-0 overflow-auto pt-1">
	{@render children()}
</div>