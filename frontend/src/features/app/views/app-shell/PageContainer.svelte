<script lang="ts">
	import type { Snippet } from "svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";

	type Props = {
		children: Snippet;
	};
	const { children }: Props = $props();
</script>

<div class="w-full max-w-full h-full max-h-full min-h-0 overflow-hidden flex flex-col gap-2 px-2 pb-2 pt-1 border rounded-lg bg-surface-300">
	<div class="border-b flex justify-between items-bottom">
		<span class="text-xl text-surface-content/50 w-fit self-bottom flex gap-1 items-end">
			{#each appShell.breadcrumbs as c, i}
				{@const last = i === appShell.breadcrumbs.length - 1}
				{#if i > 0}
					<span>/</span>
				{/if}
		
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
			{/each}
		</span>

		{#if appShell.pageActionsComponent}
			<div class="pb-1">
				<appShell.pageActionsComponent />
			</div>
		{/if}
	</div>

	<div class="flex flex-col flex-1 min-h-0 overflow-auto">
		{@render children()}
	</div>
</div>
