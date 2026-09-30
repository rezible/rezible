<script lang="ts">
	import type { PageRelatedLink } from "$lib/app-shell.svelte";
	import { Button } from "$components/ui/button";
	import * as Popover from "$components/ui/popover";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import RiArrowRightUpLine from "remixicon-svelte/icons/arrow-right-up-line";

	type Props = { links: readonly PageRelatedLink[] };
	const { links }: Props = $props();

	const triggerLabel = $derived.by(() => {
		const kinds = new Set(links.map((link) => link.kind));
		if (kinds.size === 1) {
			return `${links.length} ${links[0].kind.toLowerCase()}s`;
		}
		return `${links.length} related`;
	});
</script>

{#snippet linkContent(link: PageRelatedLink)}
	<link.icon class="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
	<span class="sr-only">{link.kind}:</span>
	<span class="min-w-0 truncate">{link.label}</span>
	{#if link.status}
		<StatusBadge status={link.status} variant="inline" />
	{/if}
	<RiArrowRightUpLine class="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
{/snippet}

<nav aria-label="Related" class="flex items-center max-sm:hidden">
	{#if links.length === 1}
		{@const link = links[0]}
		<a
			href={link.path}
			class="inline-flex h-8 max-w-72 items-center gap-2 rounded-md border border-border bg-card px-2.5 text-sm hover:bg-accent focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
		>
			{@render linkContent(link)}
		</a>
	{:else}
		<Popover.Root>
			<Popover.Trigger>
				{#snippet child({ props })}
					<Button {...props} variant="outline" size="sm">{triggerLabel}</Button>
				{/snippet}
			</Popover.Trigger>
			<Popover.Content align="end" class="flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-1 p-2">
				{#each links as link (link.key)}
					<a
						href={link.path}
						class="flex min-w-0 items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
					>
						{@render linkContent(link)}
					</a>
				{/each}
			</Popover.Content>
		</Popover.Root>
	{/if}
</nav>
